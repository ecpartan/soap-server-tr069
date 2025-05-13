package server

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/clbanning/mxj/v2"
	"github.com/dgrijalva/lfu-go"
	dac "github.com/xinsnake/go-http-digest-auth-client"
)

type setMap map[int]struct{}

// OperationHandlerFunc runs the actual business logic - request is whatever you constructed in RequestFactoryFunc
type OperationHandlerFunc func(request interface{}, w http.ResponseWriter, httpRequest *http.Request) (response interface{}, err error)

// RequestFactoryFunc constructs a request object for OperationHandlerFunc
type RequestFactoryFunc func() interface{}

type dummyContent struct{}

type operationHandler struct {
	requestFactory RequestFactoryFunc
	handler        OperationHandlerFunc
}

type responseWriter struct {
	log           func(...interface{})
	w             http.ResponseWriter
	outputStarted bool
}

func (w *responseWriter) Header() http.Header {
	return w.w.Header()
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.outputStarted = true
	if w.log != nil {
		w.log("writing response: ", string(b))
	}
	return w.w.Write(b)
}

func (w *responseWriter) WriteHeader(code int) {
	w.w.WriteHeader(code)
}

type responseTask struct {
	respChan chan any
	serial   string
	respList []any
}

// Server a SOAP server, which can be run standalone or used as a http.HandlerFunc
type Server struct {
	Log             func(...interface{}) // do nothing on nil or add your fmt.Print* or log.*
	handlers        map[string]map[string]map[string]*operationHandler
	Marshaller      XMLMarshaller
	ContentType     string
	SoapVersion     string
	RequestModifyFn func(r *http.Request) *http.Request
	context         context.Context
	Cache           *lfu.Cache
	mapResponse     map[string](responseTask)
	wg              sync.WaitGroup
}

type DeviceId struct {
	Manufacturer string `xml:"Manufacturer"`
	OUI          string `xml:"OUI"`
	ProductClass string `xml:"ProductClass"`
	SerialNumber string `xml:"SerialNumber"`
}

type EnvInfo struct {
	XMLName xml.Name `xml:"SOAP-ENV:Envelope"`
	SOAPENV string   `xml:"xmlns:SOAP-ENV,attr"`
	SOAPENC string   `xml:"xmlns:SOAP-ENC,attr"`
	Xsi     string   `xml:"xmlns:xsi,attr"`
	Xsd     string   `xml:"xmlns:xsd,attr"`
	Cwmp    string   `xml:"xmlns:cwmp,attr"`
}

type HeaderInfo struct {
	Header struct {
		Text string `xml:",chardata"`
		ID   struct {
			Text           string `xml:",chardata"`
			MustUnderstand string `xml:"SOAP-ENV:mustUnderstand,attr"`
		} `xml:"cwmp:ID"`
	} `xml:"SOAP-ENV:Header"`
}

type InformResponse struct {
	EnvInfo
	HeaderInfo `xml:"Header"`
	Body       struct {
		Text           string `xml:",chardata"`
		InformResponse struct {
			Text         string `xml:",chardata"`
			MaxEnvelopes int    `xml:"MaxEnvelopes"`
		} `xml:"InformResponse"`
	} `xml:"SOAP-ENV:Body"`
}

type GetBody struct {
	Body struct {
		Text               string `xml:",chardata"`
		GetParameterValues struct {
			Text           string `xml:",chardata"`
			ParameterNames struct {
				Text      string   `xml:",chardata"`
				ArrayType string   `xml:"SOAP-ENC:arrayType,attr"`
				String    []string `xml:"string"`
			} `xml:"ParameterNames"`
		} `xml:"cwmp:GetParameterValues"`
	} `xml:"SOAP-ENV:Body"`
}

type setValue struct {
	Text string `xml:",chardata"`
	Type string `xml:"xsi:type,attr"`
}

type setParameterValueStruct struct {
	Text  string   `xml:",chardata"`
	Name  string   `xml:"Name"`
	Value setValue `xml:"Value"`
}

type SetBody struct {
	Body struct {
		Text               string `xml:",chardata"`
		SetParameterValues struct {
			Text          string `xml:",chardata"`
			ParameterList struct {
				Text                    string                    `xml:",chardata"`
				ArrayType               string                    `xml:"SOAP-ENC:arrayType,attr"`
				SetParameterValueStruct []setParameterValueStruct `xml:"ParameterValueStruct"`
			} `xml:"ParameterList"`
			ParameterKey string `xml:"ParameterKey"`
		} `xml:"cwmp:SetParameterValues"`
	} `xml:"SOAP-ENV:Body"`
}

type AddBody struct {
	Body struct {
		Text      string `xml:",chardata"`
		AddObject struct {
			Text         string `xml:",chardata"`
			ObjectName   string `xml:"ObjectName"`
			ParameterKey string `xml:"ParameterKey"`
		} `xml:"cwmp:AddObject"`
	} `xml:"SOAP-ENV:Body"`
}

type DeleteBody struct {
	Body struct {
		Text         string `xml:",chardata"`
		DeleteObject struct {
			Text         string `xml:",chardata"`
			ObjectName   string `xml:"ObjectName"`
			ParameterKey string `xml:"ParameterKey"`
		} `xml:"cwmp:DeleteObject"`
	} `xml:"SOAP-ENV:Body"`
}

type SetParameterValues struct {
	EnvInfo
	HeaderInfo
	SetBody
}

type GetParameterValues struct {
	EnvInfo
	HeaderInfo
	GetBody
}

type AddObject struct {
	EnvInfo
	HeaderInfo
	AddBody
}

type DeleteObject struct {
	EnvInfo
	HeaderInfo
	DeleteBody
}

type parseMapXML map[string]interface{}

func (s *Server) PrepareHeaderInfo(mp any) {

	s.Log("PrepareHeaderInfo")

	envinfo := EnvInfo{}

	if mp != nil {
		soap_env_obj := GetXMLValue(mp, "xmlns:SOAP-ENV")
		soap_env := GetXMLValue(soap_env_obj, "#text").(string)
		if soap_env != "" {
			envinfo.SOAPENV = soap_env
		}

		soap_enc_obj := GetXMLValue(mp, "xmlns:SOAP-ENC")
		soap_enc := GetXMLValue(soap_enc_obj, "#text").(string)
		s.log("soap_env", soap_env)
		if soap_enc != "" {
			envinfo.SOAPENC = soap_enc
		}

		cwmp_obj := GetXMLValue(mp, "xmlns:cwmp")
		cwmp := GetXMLValue(cwmp_obj, "#text").(string)
		if cwmp != "" {
			envinfo.Cwmp = cwmp
		}

		xsi_obj := GetXMLValue(mp, "xmlns:xsi")
		xsi := GetXMLValue(xsi_obj, "#text").(string)
		if xsi != "" {
			envinfo.Xsi = xsi
		}

		xsd_obj := GetXMLValue(mp, "xmlns:xsd")
		xsd := GetXMLValue(xsd_obj, "#text").(string)
		if xsd != "" {
			envinfo.Xsd = xsd
		}
	} else {
		envinfo.SOAPENV = string(bNamespaceSoap12)
		envinfo.SOAPENC = string(bNamespaceEnc)
		envinfo.Cwmp = string(bNamespaceCwmp)
		envinfo.Xsd = string(bNamespaceXsd)
		envinfo.Xsi = string(bNamespaceXsi)
	}

	s.context = context.WithValue(s.context, "EnvInfo", envinfo)

}

// map[xmlns:SOAP-ENC:map[#seq:1 #text:http://schemas.xmlsoap.org/soap/encoding/]
//
//	xmlns:SOAP-ENV:map[#seq:0 #text:http://schemas.xmlsoap.org/soap/envelope/]
//	xmlns:cwmp:map[#seq:4 #text:urn:dslforum-org:cwmp-1-0]
//	xmlns:xsd:map[#seq:3 #text:http://www.w3.org/2001/XMLSchema]
//	xmlns:xsi:map[#seq:2 #text:http://www.w3.org/2001/XMLSchema-instance]]
func (s *Server) NewInformResponse(mp interface{}) *InformResponse {

	resp := &InformResponse{}
	envInfo, ok := s.context.Value("EnvInfo").(EnvInfo)
	if !ok {
		s.Log("NewInformResponse failed")
		return nil
	}

	resp.SOAPENV = envInfo.SOAPENV
	resp.SOAPENC = envInfo.SOAPENC
	resp.Cwmp = envInfo.Cwmp
	resp.Xsd = envInfo.Xsd
	resp.Xsi = envInfo.Xsi

	resp.Header.ID.MustUnderstand = "1"
	resp.Body.InformResponse.MaxEnvelopes = 1

	return resp
}

func (s *Server) NewGetParameterValues(paramlist GetParamTask) *GetParameterValues {
	s.log("NewGetParameterValues")
	resp := &GetParameterValues{}

	envInfo, ok := s.context.Value("EnvInfo").(EnvInfo)
	if !ok {
		s.Log("NewGetParameterValues failed")
		return nil
	}

	resp.SOAPENV = envInfo.SOAPENV
	resp.SOAPENC = envInfo.SOAPENC
	resp.Cwmp = envInfo.Cwmp
	resp.Xsd = envInfo.Xsd
	resp.Xsi = envInfo.Xsi

	/*
		getnames := make([]string, len(paramlist.Name))
		for i, name := range paramlist.Name {
			getnames[i] = name
		}*/
	resp.Body.GetParameterValues.ParameterNames.String = paramlist.Name
	resp.Body.GetParameterValues.ParameterNames.ArrayType = "xsd:string[" + strconv.Itoa(len(paramlist.Name)) + "]"

	resp.Header.ID.MustUnderstand = "1"

	return resp
}

func (s *Server) NewSetParameterValues(paramlist []SetParamTask) *SetParameterValues {

	resp := &SetParameterValues{}

	envInfo, ok := s.context.Value("EnvInfo").(EnvInfo)
	if !ok {
		s.Log("NewSetParameterValues failed")
		return nil
	}

	resp.SOAPENV = envInfo.SOAPENV
	resp.SOAPENC = envInfo.SOAPENC
	resp.Cwmp = envInfo.Cwmp
	resp.Xsd = envInfo.Xsd
	resp.Xsi = envInfo.Xsi

	paramstruct := &resp.Body.SetParameterValues.ParameterList.SetParameterValueStruct
	for _, param := range paramlist {
		*paramstruct = append(*paramstruct, setParameterValueStruct{
			Name: param.Name,
			Value: setValue{
				Text: param.Value,
				Type: param.Type,
			},
		})
	}

	s.log("paramstruct: %v", paramstruct)
	resp.Body.SetParameterValues.ParameterList.ArrayType = "xsd:ParameterValueStruct[" + strconv.Itoa(len(paramlist)) + "]"

	resp.Header.ID.MustUnderstand = "1"

	return resp
}

func (s *Server) NewAddObject(obj string) *AddObject {
	resp := &AddObject{}

	envInfo, ok := s.context.Value("EnvInfo").(EnvInfo)
	if !ok {
		s.Log("NewAddObject failed")
		return nil
	}

	resp.SOAPENV = envInfo.SOAPENV
	resp.SOAPENC = envInfo.SOAPENC
	resp.Cwmp = envInfo.Cwmp
	resp.Xsd = envInfo.Xsd
	resp.Xsi = envInfo.Xsi

	resp.Body.AddObject.ObjectName = obj
	resp.Header.ID.MustUnderstand = "1"
	s.log("resp.Body.AddObject.ObjectName: %v", resp.Body.AddObject.ObjectName)
	return resp
}

func (s *Server) NewDeleteObject(obj string) *DeleteObject {
	resp := &DeleteObject{}

	envInfo, ok := s.context.Value("EnvInfo").(EnvInfo)
	if !ok {
		s.Log("NewAddObject failed")
		return nil
	}

	resp.SOAPENV = envInfo.SOAPENV
	resp.SOAPENC = envInfo.SOAPENC
	resp.Cwmp = envInfo.Cwmp
	resp.Xsd = envInfo.Xsd
	resp.Xsi = envInfo.Xsi

	resp.Body.DeleteObject.ObjectName = obj
	resp.Header.ID.MustUnderstand = "1"

	return resp
}

// NewServer construct a new SOAP server
func NewServer() *Server {
	return &Server{
		handlers:    make(map[string]map[string]map[string]*operationHandler),
		Marshaller:  defaultMarshaller{},
		ContentType: SoapContentType11,
		SoapVersion: SoapVersion11,
		context:     context.Background(),
		Cache:       lfu.New(),
		mapResponse: make(map[string]responseTask),
	}
}

func (s *Server) log(args ...interface{}) {
	if s.Log == nil {
		return
	}
	pc, _, _, _ := runtime.Caller(0)
	s.Log(append([]interface{}{runtime.FuncForPC(pc).Name()}, args...)...)
}
func (s *Server) UseSoap11() {
	s.SoapVersion = SoapVersion11
	s.ContentType = SoapContentType11
}

func (s *Server) UseSoap12() {
	s.SoapVersion = SoapVersion12
	s.ContentType = SoapContentType12
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

func (s *Server) handleError(err error, w http.ResponseWriter) {
	// has to write a soap fault
	s.log("handling error:", err)
	responseEnvelope := &Envelope{}
	xmlBytes, xmlErr := s.Marshaller.Marshal(responseEnvelope)
	if xmlErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "could not marshal soap fault for: %s xmlError: %s\n", err, xmlErr)
		return
	}
	addSOAPHeader(w, len(xmlBytes), s.ContentType)
	w.Write(xmlBytes)
}

// WriteHeader first set the content-type header and then writes the header code.
func (s *Server) WriteHeader(w http.ResponseWriter, code int) {
	setContentType(w, s.ContentType)
	w.WriteHeader(code)
}

func setContentType(w http.ResponseWriter, contentType string) {
	w.Header().Set("Content-Type", contentType)
}

func addSOAPHeader(w http.ResponseWriter, contentLength int, contentType string) {
	setContentType(w, contentType)
	w.Header().Set("Content-Length", fmt.Sprint(contentLength))
}

func (s *Server) TransmitXMLReq(request any, w http.ResponseWriter) {
	xmlBytes, err := s.Marshaller.Marshal(request)
	// Adjust namespaces for SOAP 1.2
	if s.SoapVersion == SoapVersion12 {
		xmlBytes = replaceSoap11to12(xmlBytes)
	}
	if err != nil {
		s.handleError(fmt.Errorf("could not marshal response:: %s", err), w)
	}
	addSOAPHeader(w, len(xmlBytes), s.ContentType)
	w.Write(xmlBytes)

}

func (s *Server) TransInformResponse(w http.ResponseWriter, req any) {
	s.log("Enter TransInform", req)

	responseEnvelope := s.NewInformResponse(req)
	s.TransmitXMLReq(responseEnvelope, w)
}

func (s *Server) TransGetParameterValues(w http.ResponseWriter, req any) {
	s.log("TransGetParameterValues")
	s.log(req)
	s.log(reflect.TypeOf(req))
	if getList, ok := req.(GetParamTask); ok {
		responseEnvelope := s.NewGetParameterValues(getList)
		s.TransmitXMLReq(responseEnvelope, w)
		s.wg.Done()
	}
}

func (s *Server) TransSetParameterValues(w http.ResponseWriter, req any) {
	s.log("TransSetParameterValues")
	s.log(req)
	if setList, ok := req.([]SetParamTask); ok {

		responseEnvelope := s.NewSetParameterValues(setList)
		s.TransmitXMLReq(responseEnvelope, w)
		s.wg.Done()
	}
}

func (s *Server) TransAddObject(w http.ResponseWriter, req any) {
	s.log("TransAddObjectResponse")
	s.log(req)
	if addInst, ok := req.(AddTask); ok {
		responseEnvelope := s.NewAddObject(addInst.Name)
		s.TransmitXMLReq(responseEnvelope, w)
		s.wg.Done()
		s.log("add object success")
		//s.GetTasks(w, host)
	}
}

func (s *Server) TransDeleteObject(w http.ResponseWriter, req any) {
	s.log("TransDeleteObjectResponse")
	s.log(req)
	if DelInst, ok := req.(string); !ok {
		return
	} else {
		responseEnvelope := s.NewDeleteObject(DelInst)
		s.TransmitXMLReq(responseEnvelope, w)
		s.wg.Done()
	}
}

func GetXMLValue(xmlMap any, key string) any {
	if parseMap, ok := xmlMap.(map[string]any); !ok {
		return nil
	} else {
		if value, ok := parseMap[key]; ok {
			return value
		}
	}
	return nil
}

func GetXMLValueS(xmlMap any, key string) any {

	strs := strings.Split(key, ".")
	if len(strs) == 0 {
		return nil
	}

	if len(strs) == 1 {
		return GetXMLValue(xmlMap, strs[0])
	}

	//fmt.Println(strs)

	for i := 0; i < len(strs); i++ {
		xmlMap = GetXMLValue(xmlMap, strs[i])
	}

	//fmt.Println(xmlMap)
	return xmlMap
}

func GetXMLValueMap(xmlMap any, key string) map[string]any {
	if parseMap, ok := xmlMap.(map[string]any); !ok {
		return nil
	} else {
		if value, ok := parseMap[key]; ok {
			return value.(map[string]any)
		}
	}
	return nil
}

func (s *Server) CheckSoapType(addr string, mv map[string]interface{}) (map[string]interface{}, TaskType) {

	if mv == nil {
		return nil, ResponseEndSession
	}

	envelope := GetXMLValueS(mv, "SOAP-ENV:Envelope")

	if envelope == nil {
		s.log("envelope is not parseMapXML")
		return nil, ResponseUndefinded
	} else {
		attrs := GetXMLValueMap(envelope, "#attr")

		s.PrepareHeaderInfo(attrs)

		bod := GetXMLValue(envelope, "SOAP-ENV:Body")
		s.log("attrs: %v", bod)

		if bod != nil {
			inf := GetXMLValue(bod, "cwmp:Inform")

			if inf != nil {
				s.log("found Inform")
				serial := GetXMLValueS(inf, "DeviceId.SerialNumber.#text").(string)
				if serial != "" {
					AddDevicetoTaskList(serial, addr)
				}

				if _, ok := s.context.Value("DeviceID").(DeviceId); !ok {

					id := DeviceId{
						Manufacturer: GetXMLValueS(inf, "DeviceId.Manufacturer.#text").(string),
						OUI:          GetXMLValueS(inf, "DeviceId.OUI.#text").(string),
						ProductClass: GetXMLValueS(inf, "DeviceId.ProductClass.#text").(string),
						SerialNumber: GetXMLValueS(inf, "DeviceId.SerialNumber.#text").(string),
					}

					s.context = context.WithValue(s.context, "DeviceID", id)

				}

				return inf.(map[string]interface{}), Inform
			}

			ret := GetXMLValue(bod, "cwmp:GetParameterValuesResponse")

			if ret != nil {
				return ret.(map[string]interface{}), GetParameterValuesResponse
			}
			ret = GetXMLValue(bod, "cwmp:SetParameterValuesResponse")

			if ret != nil {
				return ret.(map[string]interface{}), SetParameterValuesResponse
			}

			ret = GetXMLValue(bod, "cwmp:AddObjectResponse")
			if ret != nil {
				return ret.(map[string]interface{}), AddObjectResponse
			}
			ret = GetXMLValue(bod, "cwmp:DeleteObjectResponse")
			if ret != nil {
				return ret.(map[string]interface{}), DeleteObjectResponse
			}
		}
	}
	return nil, ResponseUndefinded

}

func (s *Server) parseEventCode(mp map[string]interface{}) {
	codes := make(setMap)

	if mp != nil {
		events := GetXMLValueS(mp, "Event.EventStruct")
		if events == nil {
			return
		}

		s.log("events", events)
		if list_events, ok := events.(map[string]any); ok {

			for event, map_event := range list_events {
				s.log("event", map_event)
				if event == "EventCode" {

					eventCode := GetXMLValue(map_event, "#text").(string)

					switch eventCode {
					case "0 BOOTSTRAP":
						codes[0] = struct{}{}
					case "1 BOOT":
						codes[1] = struct{}{}
					case "2 PERIODIC":
						codes[2] = struct{}{}
					case "3 SCHEDULED":
						codes[3] = struct{}{}
					case "4 VALUE CHANGE":
						codes[4] = struct{}{}
					case "6 CONNECTION REQUEST":
						codes[6] = struct{}{}
					case "7 TRANSFER COMPLETE":
						codes[7] = struct{}{}
					}
				}
			}
		}
	}
	s.context = context.WithValue(s.context, "codes", codes)
}

func (s *Server) NextTask(serial, addr string) (TaskRequestType, Task) {

	s.log("NextTask")

	s.CheckNewConReqTasks(serial, addr)
	tasks := GetListTasksBySerial(serial, addr)

	if len(tasks) == 0 {
		return NoTaskRequestR, Task{}
	}

	s.log("tasks", tasks)

	eventCodes := s.context.Value("codes").(setMap)
	s.log("eventCodes", eventCodes)

	for _, task := range tasks {
		if _, ok := eventCodes[task.eventCode]; !ok {
			continue
		}
		s.log("Action", task.action)
		switch task.action {
		case "GetParmeterValues":
			{
				s.log("GetParmeterValues")
				DeleteTaskByID(serial, addr, task.id)

				return GetParameterValuesR, task
			}
		case "SetParameterValues":
			{
				s.log("SetParmeterValues")
				s.log(task)
				DeleteTaskByID(serial, addr, task.id)
				s.log(task)

				return SetParameterValuesR, task
			}
		case "AddObject":
			{
				s.log("AddObject")
				DeleteTaskByID(serial, addr, task.id)
				return AddObjectR, task
			}
		case "DeleteObject":
			{
				s.log("DeleteObject")
				DeleteTaskByID(serial, addr, task.id)
				return DeleteObjectR, task
			}
		}
	}

	return NoTaskRequestR, Task{}
}

func (s *Server) executeResponsetask(task_func func(w http.ResponseWriter, req any), task Task, host string, w http.ResponseWriter) {

	s.wg.Add(1)
	go func() {
		s.log(task)
		task_func(w, task.params)

		if s.mapResponse[host].respChan == nil {
			s.mapResponse[host] = responseTask{
				respChan: make(chan any),
				respList: make([]any, 0),
			}
		}
		ret := <-s.mapResponse[host].respChan
		s.log("executeResponsetask", ret)
		respTask := s.mapResponse[host]
		respTask.respList = append(respTask.respList, ret)
		s.mapResponse[host] = respTask
		s.log("executeResponsetask", s.mapResponse)
	}()
}

func SubstringInstance(message string, start, end byte) (bool, int, int) {

	if idx := strings.IndexByte(message, start); idx >= 0 {
		fmt.Println("idx", message[idx:])
		if idx_end := strings.IndexByte(message[idx:], end); idx_end >= 0 {
			return true, idx, idx + idx_end
		} else {
			return true, idx, idx + (idx - len(message) + 1)
		}
	}

	return false, -1, -1
}

func (s *Server) PrepareListTask(task Task, host string) {

	lst := s.mapResponse[host].respList
	s.log("PrepareListTask", lst, len(lst))
	s.log("PrepareListTask", task)
	if len(lst) <= 0 {
		return
	}

	switch task.action {
	case "SetParameterValues":
		{
			s.log("SetParmeterValues")
			task_params := task.params.([]SetParamTask)
			s.log("tasks", task_params)

			for k, v := range task_params {
				str := v.Name
				if ok, start, end := SubstringInstance(str, '#', '.'); ok {
					replacing_trim := str[start:end]
					s.log("replacing_trim", replacing_trim)
					if i, err := strconv.Atoi(replacing_trim[1:]); err == nil {
						if replace_trim, ok := lst[i].(string); ok {
							task_params[k].Name = str[:start] + replace_trim + str[end:]
							s.log("tasks", task_params)
						}
					}
				}
			}
		}
	case "AddObject":
		{
			task_params := task.params.(AddTask)
			str := task_params.Name
			if ok, start, end := SubstringInstance(str, '#', '.'); ok {
				replacing_trim := str[start:end]
				if i, err := strconv.Atoi(replacing_trim[1:]); err == nil {
					if replace_trim, ok := lst[i].(string); ok {
						task_params.Name = str[:start] + replace_trim + str[end:]
					}
				}
			}
		}
	}

}

func (s *Server) ExecuteTask(Action TaskRequestType, task Task, host string, w http.ResponseWriter) {
	s.log("ExecuteTask")

	if task.action == "" {
		return
	}

	switch Action {
	case GetParameterValuesR:
		{
			s.log("GetParameterValuesR")
			s.executeResponsetask(s.TransGetParameterValues, task, host, w)
		}
	case SetParameterValuesR:
		{
			s.log("SetParameterValuesR")
			s.PrepareListTask(task, host)
			s.executeResponsetask(s.TransSetParameterValues, task, host, w)
		}

	case AddObjectR:
		{
			s.log("AddObjectR")
			s.executeResponsetask(s.TransAddObject, task, host, w)
		}

	case DeleteObjectR:
		{
			s.log("DeleteObjectR")
			s.executeResponsetask(s.TransDeleteObject, task, host, w)
		}
	}
	s.wg.Wait()
	s.log("ExecuteTask end")
}

func execute_connection_request() (*http.Response, error) {
	dr := dac.NewRequest("", "", "GET", "http://localhost:8999/", "")
	response1, err := dr.Execute()

	return response1, err
}

func (s *Server) GetTasks(w http.ResponseWriter, host string) {
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

	taskAction, task := s.NextTask(serial, host)

	if taskAction == NoTaskRequestR {
		s.log("task is nil")
		w.WriteHeader(http.StatusNoContent)
		return
	} else {
		s.ExecuteTask(taskAction, task, host, w)
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.log("action handler threw up")

	if s.RequestModifyFn != nil {
		r = s.RequestModifyFn(r)
	}
	soapAction := r.Header.Get("SOAPAction")
	addr := r.RemoteAddr
	s.log("ServeHTTP method:", r.Method, ", path:", r.URL.Path, ",  soapAction:", soapAction)
	s.log("mapresponse", s.mapResponse)

	if r.URL.Path == "/request" {
		execute_connection_request()
		return
	}

	if r.URL.Path == "/addtask" {
		s.log("addtask")
		soapRequestBytes, err := io.ReadAll(r.Body)
		if err == nil {
			var getScript map[string]any
			err := json.Unmarshal(soapRequestBytes, &getScript)
			if err != nil {
				s.log("error", err)
				return
			}
			s.log("body_task", getScript)
			if err := s.ParseScriptToTask(getScript); err != nil {
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
		log:           s.Log,
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
			s.handleError(fmt.Errorf("could not read POST:: %s", err), w)
			return
		}

		mv, err := mxj.NewMapXmlSeq(soapRequestBytes)

		paramType := ResponseUndefinded
		var xml_body map[string]interface{}

		if err != nil {
			paramType = ResponseEndSession
		} else {
			s.log(mv)
			xml_body, paramType = s.CheckSoapType(addr, mv)
		}

		s.log("found soap type", paramType)

		switch paramType {
		case ResponseUndefinded:
			s.handleError(fmt.Errorf("unknown XML Soap Type"), w)
			return
		case ResponseEndSession:
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

			taskAction, task := s.NextTask(serial, addr)

			if taskAction == NoTaskRequestR {
				s.log("task is nil")
				w.WriteHeader(http.StatusNoContent)
				return
			} else {
				s.ExecuteTask(taskAction, task, addr, w)
			}
		case Inform:
			s.parseEventCode(xml_body)
			if !w.(*responseWriter).outputStarted {
				s.TransInformResponse(w, xml_body)
			}
		case GetParameterValuesResponse:
			s.ParseGetResponse(xml_body, addr)
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
