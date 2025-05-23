package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/clbanning/mxj/v2"
	"github.com/ecpartan/soap-server-tr069/httpserver"
	"github.com/ecpartan/soap-server-tr069/soaprpc"

	"github.com/dgrijalva/lfu-go"
	"github.com/ecpartan/soap-server-tr069/internal/devmodel"
	p "github.com/ecpartan/soap-server-tr069/internal/parsemap"
	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/soap-server-tr069/soap"
	"github.com/ecpartan/soap-server-tr069/tasks"
	dac "github.com/xinsnake/go-http-digest-auth-client"
)

type OperationHandlerFunc func(w http.ResponseWriter, httpRequest *http.Request)

type DevID struct {
	devmodel.ResponseTask
	soap.SoapResponse
}

type DevMap struct {
	sync.RWMutex
	Wg   *sync.WaitGroup
	devs map[string]*DevID
}

func NewDevMap() DevMap {
	return DevMap{
		devs: make(map[string]*DevID),
		Wg:   &sync.WaitGroup{},
	}
}

func (d *DevMap) Get(key string) *DevID {
	d.RLock()
	defer d.RUnlock()
	return d.devs[key]
}

func (d *DevMap) Set(key string, value *DevID) {
	d.Lock()
	defer d.Unlock()
	d.devs[key] = value
}

// Server a SOAP server, which can be run standalone or used as a http.HandlerFunc
type Server struct {
	handlers    map[string]*OperationHandlerFunc
	context     context.Context
	Cache       *lfu.Cache
	mapResponse DevMap
}

// NewServer construct a new SOAP server
func NewServer() *Server {
	return &Server{
		handlers:    make(map[string]*OperationHandlerFunc),
		context:     context.Background(),
		Cache:       lfu.New(),
		mapResponse: NewDevMap(),
	}
}

// RegisterHandler register to handle an operation. This function must not be
// called after the server has been started.
func (s *Server) RegisterHandler(path string, operationHandlerFunc OperationHandlerFunc) {
	if s.handlers == nil {
		panic("RegisterHandler called on a server with no handlers")
	}
	s.handlers[path] = &operationHandlerFunc
}

func (s *Server) execute_connection_request(serial string) (*http.Response, error) {

	logger.LogDebug("execute_connection_request", serial)

	mp := s.Cache.Get(serial)
	if mp == nil {
		logger.LogDebug("mp is nil")
		return nil, errors.New("no found in cache ")
	}
	if crURL, ok := p.GetXMLValueS(mp, "InternetGatewayDevice.ManagementServer.ConnectionRequestURL").(string); ok {
		logger.LogDebug("crURL", crURL)
		dr := dac.NewRequest("", "", "GET", crURL, "")
		response1, err := dr.Execute()

		if err != nil {
			logger.LogDebug("error in execute_connection_request")
			return response1, err
		}
	}
	return nil, errors.New("no found addres for this device by SN")
}

func (s *Server) ParseXML(addr string, mv map[string]any) soap.TaskType {
	logger.LogDebug("ParseXML", mv)
	if mv == nil {
		return soap.ResponseUndefinded
	}
	envelope := p.GetXMLValueS(mv, "SOAP-ENV:Envelope")

	if envelope == nil {
		logger.LogDebug("envelope is not parseMapXML")
		return soap.ResponseUndefinded
	} else {

		mp := s.mapResponse.Get(addr)

		logger.LogDebug("mapresponse", mp)
		mp.Env = soap.PrepareHeaderInfo(envelope)

		xml_body := p.GetXMLValue(envelope, "SOAP-ENV:Body")

		if xml_body == nil {
			return soap.ResponseUndefinded
		}
		var status = soap.ResponseUndefinded
		inf := p.GetXMLValue(xml_body, "cwmp:Inform")

		if inf != nil {
			logger.LogDebug("found Inform")
			serial := p.GetXMLValueS(inf, "DeviceId.SerialNumber.#text").(string)
			if serial == "" {
				return soap.ResponseUndefinded
			}
			tasks.AddDevicetoTaskList(serial)

			paramlist := p.GetXMLValueS(inf, "ParameterList.ParameterValueStruct").([]any)
			logger.LogDebug("paramlist", paramlist)
			tasks.UpdateCacheBySerial(serial, paramlist, s.Cache)

			mp.Body = inf.(map[string]any)
			mp.Serial = serial
			status = soap.Inform
		} else if ret, ok := p.GetXMLValue(xml_body, "cwmp:GetParameterValuesResponse").(map[string]any); ok {
			mp.Body = ret
			status = soap.GetParameterValuesResponse
		} else if ret, ok := p.GetXMLValue(xml_body, "cwmp:SetParameterValuesResponse").(map[string]any); ok {
			mp.Body = ret
			status = soap.SetParameterValuesResponse
		} else if ret, ok := p.GetXMLValue(xml_body, "cwmp:AddObjectResponse").(map[string]any); ok {
			mp.Body = ret
			status = soap.AddObjectResponse
		} else if ret, ok := p.GetXMLValue(xml_body, "cwmp:DeleteObjectResponse").(map[string]any); ok {
			mp.Body = ret
			status = soap.DeleteObjectResponse
		}

		s.mapResponse.Set(addr, mp)

		return status
	}
}

func (s *Server) PerformConReq(w http.ResponseWriter, r *http.Request) {
	logger.LogDebug("addtask")
	soapRequestBytes, err := io.ReadAll(r.Body)
	if err == nil {
		var getScript map[string]any
		err := json.Unmarshal(soapRequestBytes, &getScript)
		if err != nil {
			logger.LogDebug("error", err)
			return
		}
		logger.LogDebug("body_task", getScript)

		var serial string
		serial, err = tasks.ParseScriptToTask(getScript)
		if err != nil || serial == "" {
			logger.LogDebug("error", err)
			return
		}
		if _, err := s.execute_connection_request(serial); err != nil {
			logger.LogDebug("error", err)
			return
		}

		return
	}

}

func (s *Server) MainHandler(w http.ResponseWriter, r *http.Request) {

	soapRequestBytes, err := io.ReadAll(r.Body)

	if err != nil {
		httpserver.HandleError(fmt.Errorf("could not read POST:: %s", err), w)
		return
	}
	addr := r.RemoteAddr

	var mv map[string]any

	mv, err = mxj.NewMapXmlSeq(soapRequestBytes)

	paramType := soap.ResponseUndefinded

	mp := s.mapResponse.Get(addr)
	if mp == nil {
		s.mapResponse.Set(addr, &DevID{
			SoapResponse: soap.InitSoapResponse(),
			ResponseTask: devmodel.InitResponseTask(),
		})
		mp = s.mapResponse.Get(addr)
	}

	if err != nil || mv == nil {
		logger.LogDebug("End session")
		if tasks.GetTasks(w, addr, &mp.ResponseTask, &mp.SoapResponse, s.mapResponse.Wg) {
			return
		}

	} else {
		paramType = s.ParseXML(addr, mv)
		logger.LogDebug("mapresponse", s.mapResponse)

		xml_body := mp.Body

		logger.LogDebug("found soap type", paramType, xml_body)

		switch paramType {
		case soap.ResponseUndefinded:
			httpserver.HandleError(fmt.Errorf("unknown XML Soap Type"), w)
			return
		case soap.Inform:
			if !w.(*httpserver.ResponseWriter).OutputStarted {
				soaprpc.TransInformResponse(w, mp.ResponseTask.Body, &mp.SoapResponse)
			}
		case soap.GetParameterValuesResponse:
			tasks.ParseGetResponse(mp.ResponseTask.Body, mp.Serial, mp.RespChan, s.Cache)
			logger.LogDebug("GetParameterValuesResponse")
			if tasks.GetTasks(w, addr, &mp.ResponseTask, &mp.SoapResponse, s.mapResponse.Wg) {
				return
			}
		case soap.SetParameterValuesResponse:

			tasks.ParseSetResponse(xml_body, mp.RespChan)
			logger.LogDebug("SetParameterValuesResponse")
			if tasks.GetTasks(w, addr, &mp.ResponseTask, &mp.SoapResponse, s.mapResponse.Wg) {
				return
			}
		case soap.AddObjectResponse:
			tasks.ParseAddResponse(mp.Body, mp.RespChan)
			logger.LogDebug("AddObjectResponse")
			tasks.GetTasks(w, addr, &mp.ResponseTask, &mp.SoapResponse, s.mapResponse.Wg)

		case soap.DeleteObjectResponse:
			tasks.ParseDeleteResponse(xml_body, mp.RespChan)
			logger.LogDebug("DeleteObjectResponse")
			if tasks.GetTasks(w, addr, &mp.ResponseTask, &mp.SoapResponse, s.mapResponse.Wg) {
				return
			}
		default:
			break
		}
	}
}

func (s *Server) checkHandlers(w http.ResponseWriter, r *http.Request) bool {
	if pathHandlers, ok := s.handlers[r.URL.Path]; ok {
		(*pathHandlers)(w, r)
		return true
	}
	return false
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	logger.LogDebug("ServeHTTP enter:", r.Method, ", path:", r.URL.Path)

	w = &httpserver.ResponseWriter{W: w, OutputStarted: false}

	switch r.Method {
	case "POST":
		if s.checkHandlers(w, r) {
			return
		}

	default:
		httpserver.HandleError(errors.New("this is a soap service - you have to POST soap requests"), w)
	}
}
