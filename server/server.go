package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/ecpartan/soap-server-tr069/httpserver"

	"github.com/clbanning/mxj/v2"
	"github.com/dgrijalva/lfu-go"
	"github.com/ecpartan/soap-server-tr069/internal/devmodel"
	p "github.com/ecpartan/soap-server-tr069/internal/parsemap"
	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/soap-server-tr069/soap"
	"github.com/ecpartan/soap-server-tr069/soaprpc"
	"github.com/ecpartan/soap-server-tr069/tasks"
	dac "github.com/xinsnake/go-http-digest-auth-client"
)

// OperationHandlerFunc runs the actual business logic - request is whatever you constructed in RequestFactoryFunc
type OperationHandlerFunc func(request interface{}, w http.ResponseWriter, httpRequest *http.Request) (response interface{}, err error)

// RequestFactoryFunc constructs a request object for OperationHandlerFunc
type RequestFactoryFunc func() interface{}

type dummyContent struct{}

type operationHandler struct {
	requestFactory RequestFactoryFunc
	handler        OperationHandlerFunc
}

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
	handlers        map[string]map[string]map[string]*operationHandler
	RequestModifyFn func(r *http.Request) *http.Request
	context         context.Context
	Cache           *lfu.Cache
	mapResponse     DevMap
}

// NewServer construct a new SOAP server
func NewServer() *Server {
	return &Server{
		handlers:    make(map[string]map[string]map[string]*operationHandler),
		context:     context.Background(),
		Cache:       lfu.New(),
		mapResponse: NewDevMap(),
	}
}

// RegisterHandler register to handle an operation. This function must not be
// called after the server has been started.
func (s *Server) RegisterHandler(path string, action string, messageType string, requestFactory RequestFactoryFunc, operationHandlerFunc OperationHandlerFunc) {
	if _, ok := s.handlers[path]; !ok {
		s.handlers[path] = make(map[string]map[string]*operationHandler)
	}

	if _, ok := s.handlers[path][action]; !ok {
		s.handlers[path][action] = make(map[string]*operationHandler)
	}
	s.handlers[path][action][messageType] = &operationHandler{
		handler:        operationHandlerFunc,
		requestFactory: requestFactory,
	}
}

func execute_connection_request() (*http.Response, error) {
	dr := dac.NewRequest("", "", "GET", "http://localhost:8999/", "")
	response1, err := dr.Execute()
	if err != nil {
		logger.LogDebug("error in execute_connection_request")
		return nil, err
	}
	return response1, err
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

		bod := p.GetXMLValue(envelope, "SOAP-ENV:Body")

		if bod == nil {
			return soap.ResponseUndefinded
		}
		var status = soap.ResponseUndefinded
		inf := p.GetXMLValue(bod, "cwmp:Inform")

		if inf != nil {
			logger.LogDebug("found Inform")
			serial := p.GetXMLValueS(inf, "DeviceId.SerialNumber.#text").(string)
			if serial == "" {
				return soap.ResponseUndefinded
			}
			tasks.AddDevicetoTaskList(serial, addr)

			mp.Body = inf.(map[string]any)
			mp.Serial = serial
			status = soap.Inform
		} else if ret, ok := p.GetXMLValue(bod, "cwmp:GetParameterValuesResponse").(map[string]any); ok {
			mp.Body = ret
			status = soap.GetParameterValuesResponse
		} else if ret, ok := p.GetXMLValue(bod, "cwmp:SetParameterValuesResponse").(map[string]any); ok {
			mp.Body = ret
			status = soap.SetParameterValuesResponse
		} else if ret, ok := p.GetXMLValue(bod, "cwmp:AddObjectResponse").(map[string]any); ok {
			mp.Body = ret
			status = soap.AddObjectResponse
		} else if ret, ok := p.GetXMLValue(bod, "cwmp:DeleteObjectResponse").(map[string]any); ok {
			mp.Body = ret
			status = soap.DeleteObjectResponse
		}

		s.mapResponse.Set(addr, mp)

		return status
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger.LogDebug("action handler threw up")

	if s.RequestModifyFn != nil {
		r = s.RequestModifyFn(r)
	}
	soapAction := r.Header.Get("SOAPAction")
	addr := r.RemoteAddr
	logger.LogDebug("ServeHTTP method:", r.Method, ", path:", r.URL.Path, ",  soapAction:", soapAction)

	if r.URL.Path == "/request" {
		execute_connection_request()
		return
	}

	if r.URL.Path == "/addtask" {
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
			if err := tasks.ParseScriptToTask(getScript); err != nil {
				logger.LogDebug("error", err)
				return
			}
			execute_connection_request()

			return
		}
		return
	}
	// we have a valid request time to call the handler
	w = &httpserver.ResponseWriter{W: w, OutputStarted: false}

	switch r.Method {
	case "POST":

		soapRequestBytes, err := io.ReadAll(r.Body)

		// Our structs for Envelope, Header, Body and Fault are tagged with namespace for SOAP 1.1
		// Therefore we must adjust namespaces for incoming SOAP 1.2 messages
		/*if s.SoapVersion == SoapVersion12 {
			soapRequestBytes = replaceSoap12to11(soapRequestBytes)
		}
		*/
		if err != nil {
			httpserver.HandleError(fmt.Errorf("could not read POST:: %s", err), w)
			return
		}

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

	default:
		// this will be a soap fault !?
		httpserver.HandleError(errors.New("this is a soap service - you have to POST soap requests"), w)

	}
}
