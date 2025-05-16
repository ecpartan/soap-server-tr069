package soaprpc

import (
	"net/http"

	"github.com/ecpartan/soap-server-tr069/httpserver"
	"github.com/ecpartan/soap-server-tr069/internal/taskmodel"
	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/soap-server-tr069/soap"
)

func TransInformResponse(w http.ResponseWriter, xml_body map[string]any, sp *soap.SoapResponse) {

	logger.LogDebug("Enter TransInform")
	sp.EventCodes = soap.ParseEventCode(xml_body)

	responseEnvelope := soap.NewInformResponse(sp.Env)
	httpserver.TransmitXMLReq(responseEnvelope, w, sp.ContentType)
}

func TransGetParameterValues(w http.ResponseWriter, req any, sp *soap.SoapResponse) {

	logger.LogDebug("TransGetParameterValues")

	if getList, ok := req.(taskmodel.GetParamTask); ok {
		responseEnvelope := soap.NewGetParameterValues(getList, sp.Env)
		httpserver.TransmitXMLReq(responseEnvelope, w, sp.ContentType)
	}
}

func TransSetParameterValues(w http.ResponseWriter, req any, sp *soap.SoapResponse) {
	logger.LogDebug("TransSetParameterValues")

	if setList, ok := req.([]taskmodel.SetParamTask); ok {
		responseEnvelope := soap.NewSetParameterValues(setList, sp.Env)
		httpserver.TransmitXMLReq(responseEnvelope, w, sp.ContentType)
	}
}

func TransAddObject(w http.ResponseWriter, req any, sp *soap.SoapResponse) {
	logger.LogDebug("TransAddObjectResponse")

	if addInst, ok := req.(taskmodel.AddTask); ok {
		responseEnvelope := soap.NewAddObject(addInst.Name, sp.Env)
		httpserver.TransmitXMLReq(responseEnvelope, w, sp.ContentType)
	}
}

func TransDeleteObject(w http.ResponseWriter, req any, sp *soap.SoapResponse) {
	logger.LogDebug("TransDeleteObjectResponse")

	if DelInst, ok := req.(string); ok {
		responseEnvelope := soap.NewDeleteObject(DelInst, sp.Env)
		httpserver.TransmitXMLReq(responseEnvelope, w, sp.ContentType)
	}
}
