package httpserver

import (
	"net/http"
	"reflect"

	"github.com/ecpartan/soap-server-tr069/internal/taskmodel"
	logger "github.com/ecpartan/soap-server-tr069/log"

	"github.com/ecpartan/soap-server-tr069/soap"
)

func TransInformResponse(w http.ResponseWriter, xml_body map[string]any, sp *soap.SoapSessionInfo) {

	logger.LogDebug("Enter TransInform")
	sp.EventCodes = soap.ParseEventCode(xml_body)

	responseEnvelope := soap.NewInformResponse(sp.Env)
	TransmitXMLReq(responseEnvelope, w, sp.ContentType, sp.AuthUsername, sp.AuthPassword)
	logger.LogDebug("end")
}

func TransGetParameterValues(w http.ResponseWriter, req any, sp *soap.SoapSessionInfo) {

	logger.LogDebug("TransGetParameterValues")

	if getList, ok := req.(taskmodel.GetParamValTask); ok {
		responseEnvelope := soap.NewGetParameterValues(getList, sp.Env)
		TransmitXMLReq(responseEnvelope, w, sp.ContentType, sp.AuthUsername, sp.AuthPassword)
	}
}

func TransSetParameterValues(w http.ResponseWriter, req any, sp *soap.SoapSessionInfo) {
	logger.LogDebug("TransSetParameterValues")

	if setList, ok := req.([]taskmodel.SetParamValTask); ok {
		responseEnvelope := soap.NewSetParameterValues(setList, sp.Env)
		TransmitXMLReq(responseEnvelope, w, sp.ContentType, sp.AuthUsername, sp.AuthPassword)
	}
}

func TransAddObject(w http.ResponseWriter, req any, sp *soap.SoapSessionInfo) {
	logger.LogDebug("TransAddObjectResponse")

	if addInst, ok := req.(taskmodel.AddTask); ok {
		responseEnvelope := soap.NewAddObject(addInst.Name, sp.Env)
		TransmitXMLReq(responseEnvelope, w, sp.ContentType, sp.AuthUsername, sp.AuthPassword)
	}
}

func TransDeleteObject(w http.ResponseWriter, req any, sp *soap.SoapSessionInfo) {
	logger.LogDebug("TransDeleteObjectResponse")

	if DelInst, ok := req.(string); ok {
		responseEnvelope := soap.NewDeleteObject(DelInst, sp.Env)
		TransmitXMLReq(responseEnvelope, w, sp.ContentType, sp.AuthUsername, sp.AuthPassword)
	}
}

func TransGetParameterNames(w http.ResponseWriter, req any, sp *soap.SoapSessionInfo) {
	logger.LogDebug("TransGetParameterNames")

	if getlist, ok := req.(taskmodel.GetParamNamesTask); ok {
		responseEnvelope := soap.NewGetParameterNames(getlist, sp.Env)
		TransmitXMLReq(responseEnvelope, w, sp.ContentType, sp.AuthUsername, sp.AuthPassword)
	}
}

func TransGetParameterAttributes(w http.ResponseWriter, req any, sp *soap.SoapSessionInfo) {
	logger.LogDebug("TransGetParameterAttributes")

	if getlist, ok := req.(taskmodel.GetParamAttrTask); ok {
		responseEnvelope := soap.NewGetParameterAttributes(getlist, sp.Env)
		TransmitXMLReq(responseEnvelope, w, sp.ContentType, sp.AuthUsername, sp.AuthPassword)
	}
}

func TransSetParameterAttributes(w http.ResponseWriter, req any, sp *soap.SoapSessionInfo) {
	logger.LogDebug("TransSetParameterAttributes")

	if setlist, ok := req.([]taskmodel.SetParamAttrTask); ok {
		responseEnvelope := soap.NewSetParameterAttributes(setlist, sp.Env)
		TransmitXMLReq(responseEnvelope, w, sp.ContentType, sp.AuthUsername, sp.AuthPassword)
	}
}

func TransReboot(w http.ResponseWriter, req any, sp *soap.SoapSessionInfo) {
	logger.LogDebug("TransReboot")

	responseEnvelope := soap.NewReboot(sp.Env)
	TransmitXMLReq(responseEnvelope, w, sp.ContentType, sp.AuthUsername, sp.AuthPassword)
}

func TransFactoryReset(w http.ResponseWriter, req any, sp *soap.SoapSessionInfo) {
	logger.LogDebug("TransFactoryReset")

	responseEnvelope := soap.NewFactoryReset(sp.Env)
	TransmitXMLReq(responseEnvelope, w, sp.ContentType, sp.AuthUsername, sp.AuthPassword)
}

func TransDownload(w http.ResponseWriter, req any, sp *soap.SoapSessionInfo) {
	logger.LogDebug("TransDownload")

	if download, ok := req.(taskmodel.DownloadTask); ok {
		responseEnvelope := soap.NewDownload(download, sp.Env)
		TransmitXMLReq(responseEnvelope, w, sp.ContentType, sp.AuthUsername, sp.AuthPassword)
	}
}

func TransUpload(w http.ResponseWriter, req any, sp *soap.SoapSessionInfo) {
	logger.LogDebug("TransUpload", req, reflect.TypeOf(req))

	if upload, ok := req.(taskmodel.UploadTask); ok {
		responseEnvelope := soap.NewUpload(upload, sp.Env)
		TransmitXMLReq(responseEnvelope, w, sp.ContentType, sp.AuthUsername, sp.AuthPassword)
	}
}

func TransGetRPCMethods(w http.ResponseWriter, req any, sp *soap.SoapSessionInfo) {
	logger.LogDebug("TransGetRPCMethods")

	responseEnvelope := soap.NewGetRPCMethods(sp.Env)
	TransmitXMLReq(responseEnvelope, w, sp.ContentType, sp.AuthUsername, sp.AuthPassword)
}

func TransTransferCompleteResponse(w http.ResponseWriter, req any, sp *soap.SoapSessionInfo) {
	logger.LogDebug("TransTransferCompleteResponse")

	responseEnvelope := soap.NewTransferCompleteResponse(sp.Env)
	TransmitXMLReq(responseEnvelope, w, sp.ContentType, sp.AuthUsername, sp.AuthPassword)
}
