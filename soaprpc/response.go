package response

import (
	"net/http"

	"github.com/ecpartan/soap-server-tr069/httpserver"
	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/soap-server-tr069/tasks"
)

func TransInformResponse(w http.ResponseWriter, xml_body map[string]any, soaptask *SoapResponse) {

	logger.LogDebug("Enter TransInform")
	soaptask.EventCodes = ParseEventCode(xml_body)

	responseEnvelope := NewInformResponse(soaptask.Env)
	httpserver.TransmitXMLReq(responseEnvelope, w, soaptask.ContentType)
}

func TransGetParameterValues(w http.ResponseWriter, xml_body any, soaptask *SoapResponse) {

	logger.LogDebug("TransGetParameterValues")

	if getList, ok := xml_body.(tasks.GetParamTask); ok {
		responseEnvelope := NewGetParameterValues(getList, soaptask.Env)
		httpserver.TransmitXMLReq(responseEnvelope, w, soaptask.ContentType)
		soaptask.wg.Done()
	}
}

func TransSetParameterValues(w http.ResponseWriter, req any) {
	logger.LogDebug("TransSetParameterValues")

	if setList, ok := req.([]tasks.SetParamTask); ok {

		responseEnvelope := NewSetParameterValues(setList)
		s.TransmitXMLReq(responseEnvelope, w)
		s.wg.Done()
	}
}

func TransAddObject(w http.ResponseWriter, req any) {
	logger.LogDebug("TransAddObjectResponse")
	if addInst, ok := req.(tasks.AddTask); ok {
		responseEnvelope := NewAddObject(addInst.Name)
		s.TransmitXMLReq(responseEnvelope, w)
		s.wg.Done()
		logger.LogDebug("add object success")
	}
}

func TransDeleteObject(w http.ResponseWriter, req any) {
	logger.LogDebug("TransDeleteObjectResponse")
	if DelInst, ok := req.(string); !ok {
		return
	} else {
		responseEnvelope := NewDeleteObject(DelInst)
		s.TransmitXMLReq(responseEnvelope, w)
		s.wg.Done()
	}
}
