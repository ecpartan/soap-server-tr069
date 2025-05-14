package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"soap-server-tr069/httpserver"
	"sync"

	"github.com/clbanning/mxj/v2"
	"github.com/dgrijalva/lfu-go"
	_ "github.com/ecpartan/soap-server-tr069/internal/xml"
	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/soap-server-tr069/soap"
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

type responseTask struct {
	respChan    chan any
	serial      string
	respList    []any
	ContentType string
	SoapVersion string
}

// Server a SOAP server, which can be run standalone or used as a http.HandlerFunc
type Server struct {
	handlers        map[string]map[string]map[string]*operationHandler
	Marshaller      XMLMarshaller
	RequestModifyFn func(r *http.Request) *http.Request
	context         context.Context
	Cache           *lfu.Cache
	mapResponse     map[string](responseTask)
	wg              sync.WaitGroup
}

// NewServer construct a new SOAP server
func NewServer() *Server {
	return &Server{
		handlers:    make(map[string]map[string]map[string]*operationHandler),
		Marshaller:  defaultMarshaller{},
		context:     context.Background(),
		Cache:       lfu.New(),
		mapResponse: make(map[string]responseTask),
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

func (s *Server) TransInformResponse(w http.ResponseWriter, req any) {

	logger.LogDebug("Enter TransInform", req)

	responseEnvelope := soap.NewInformResponse(s.context, req)
	s.TransmitXMLReq(responseEnvelope, w)
}

func (s *Server) TransGetParameterValues(w http.ResponseWriter, req any) {

	logger.LogDebug("TransGetParameterValues")

	if getList, ok := req.(tasks.GetParamTask); ok {
		responseEnvelope := soap.NewGetParameterValues(s.context, getList)
		s.TransmitXMLReq(responseEnvelope, w)
		s.wg.Done()
	}
}

func (s *Server) TransSetParameterValues(w http.ResponseWriter, req any) {
	logger.LogDebug("TransSetParameterValues")

	if setList, ok := req.([]tasks.SetParamTask); ok {

		responseEnvelope := soap.NewSetParameterValues(setList)
		s.TransmitXMLReq(responseEnvelope, w)
		s.wg.Done()
	}
}

func (s *Server) TransAddObject(w http.ResponseWriter, req any) {
	logger.LogDebug("TransAddObjectResponse")
	if addInst, ok := req.(tasks.AddTask); ok {
		responseEnvelope := soap.NewAddObject(addInst.Name)
		s.TransmitXMLReq(responseEnvelope, w)
		s.wg.Done()
		logger.LogDebug("add object success")
	}
}

func (s *Server) TransDeleteObject(w http.ResponseWriter, req any) {
	logger.LogDebug("TransDeleteObjectResponse")
	if DelInst, ok := req.(string); !ok {
		return
	} else {
		responseEnvelope := soap.NewDeleteObject(DelInst)
		s.TransmitXMLReq(responseEnvelope, w)
		s.wg.Done()
	}
}

func execute_connection_request() (*http.Response, error) {
	dr := dac.NewRequest("", "", "GET", "http://localhost:8999/", "")
	response1, err := dr.Execute()

	return response1, err
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger.LogDebug("action handler threw up")

	if s.RequestModifyFn != nil {
		r = s.RequestModifyFn(r)
	}
	soapAction := r.Header.Get("SOAPAction")
	addr := r.RemoteAddr
	logger.LogDebug("ServeHTTP method:", r.Method, ", path:", r.URL.Path, ",  soapAction:", soapAction)
	logger.LogDebug("mapresponse", s.mapResponse)

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
				s.log("error", err)
				return
			}
			execute_connection_request()

			return
		}
		return
	}
	// we have a valid request time to call the handler
	w = &responseWriter{
		w:             w,
		outputStarted: false,
	}
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

		mv, err := mxj.NewMapXmlSeq(soapRequestBytes)

		paramType := tasks.ResponseUndefinded
		var xml_body map[string]interface{}

		if err != nil {
			paramType = tasks.ResponseEndSession
		} else {
			xml_body, paramType = soap.CheckSoapType(addr, mv)
		}

		s.log("found soap type", paramType)

		switch paramType {
		case tasks.ResponseUndefinded:
			s.handleError(fmt.Errorf("unknown XML Soap Type"), w)
			return
		case tasks.ResponseEndSession:
			var serial string
			if deviceID, ok := s.context.Value("DeviceID").(DeviceId); ok {
				s.log("deviceID", deviceID)
				serial = deviceID.SerialNumber
			}
			if serial == "" {
				s.log("serial is empty")
				w.WriteHeader(http.StatusNoContent)
				return
			}

			taskAction, task := tasks.NextTask(serial, addr)

			if taskAction == tasks.NoTaskRequestR {
				s.log("task is nil")
				w.WriteHeader(http.StatusNoContent)
				return
			} else {
				tasks.ExecuteTask(taskAction, task, addr, w)
			}
		case tasks.Inform:
			soap.ParseEventCode(xml_body)
			if !w.(*responseWriter).outputStarted {
				s.TransInformResponse(w, xml_body)
			}
		case tasks.GetParameterValuesResponse:
			task.s.ParseGetResponse(xml_body, addr)
			s.log("GetParameterValuesResponse")
			s.GetTasks(w, addr)
		case SetParameterValuesResponse:
			s.ParseSetResponse(xml_body, addr)
			s.log("SetParameterValuesResponse")
			s.GetTasks(w, addr)
		case AddObjectResponse:
			s.ParseAddResponse(xml_body, addr)
			s.log("AddObjectResponse")
			s.GetTasks(w, addr)
		case DeleteObjectResponse:
			s.ParseDeleteResponse(xml_body, addr)
			s.log("DeleteObjectResponse")
			s.GetTasks(w, addr)
		default:
			break
		}

		/*
			actionHandler, ok := actionHandlers[""]

			if !ok {
				s.handleError(fmt.Errorf("no action handler for content type: %q", "envel"), w)
				return
			}

			request := actionHandler.requestFactory()

			response, err := actionHandler.handler(request, w, r)
			if err != nil {
				s.log("action handler threw up")
				s.handleError(err, w)
				return
			}

			s.log("result", s.jsonDump(response))*/

	default:
		// this will be a soap fault !?
		s.handleError(errors.New("this is a soap service - you have to POST soap requests"), w)
	}
}
func (s *Server) jsonDump(v interface{}) string {
	if s.Log == nil {
		return "not dumping"
	}
	jsonBytes, err := json.MarshalIndent(v, "", "	")
	if err != nil {
		return "error in json dump :: " + err.Error()
	}
	return string(jsonBytes)
}
