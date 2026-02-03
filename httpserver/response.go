package httpserver

import (
	"net/http"
	"reflect"

	"github.com/ecpartan/soap-server-tr069/internal/devmap"
	"github.com/ecpartan/soap-server-tr069/internal/taskmodel"
	logger "github.com/ecpartan/soap-server-tr069/log"

	"github.com/ecpartan/soap-server-tr069/soap"
)

type TransViewInfo struct {
	sp     soap.EnvInfo
	devrun devmap.DevRun
	sn     string
}

func TransInformResponse(w http.ResponseWriter, sp *soap.SoapSessionInfo, devrun *devmap.DevRun) {

	logger.LogDebug("Enter TransInform")
	responseEnvelope := soap.NewInformResponse(sp.Env)
	TransmitXMLReq(responseEnvelope, w, sp.ContentType, devrun.AuthUsername, devrun.AuthPassword)
	logger.LogDebug("end")
}

func TransGetParameterValues(w http.ResponseWriter, req any, sp *soap.SoapSessionInfo, devrun *devmap.DevRun) {

	logger.LogDebug("TransGetParameterValues")

	if getList, ok := req.(taskmodel.GetParamValTask); ok {
		responseEnvelope := soap.NewGetParameterValues(getList, sp.Env)
		TransmitXMLReq(responseEnvelope, w, sp.ContentType, devrun.AuthUsername, devrun.AuthPassword)
	}
}

func updateSessionInfo(devrun *devmap.DevRun, req []taskmodel.SetParamValTask) {
	logger.LogDebug("updateSessionInfo", req)
	for _, setList := range req {
		switch setList.Name {
		case soap.CR_U:
			devrun.ConnectionRequestUsername = setList.Value
		case soap.CR_P:
			devrun.ConnectionRequestPassword = setList.Value
		case soap.A_AUTHU:
			devrun.AuthUsername = setList.Value
		case soap.A_AUTHP:
			devrun.AuthPassword = setList.Value
		}
	}

}

func TransSetParameterValues(w http.ResponseWriter, req any, sp *soap.SoapSessionInfo, devrun *devmap.DevRun) {
	logger.LogDebug("TransSetParameterValues")

	if setList, ok := req.([]taskmodel.SetParamValTask); ok {
		updateSessionInfo(devrun, setList)

		responseEnvelope := soap.NewSetParameterValues(setList, sp.Env)
		TransmitXMLReq(responseEnvelope, w, sp.ContentType, devrun.AuthUsername, devrun.AuthPassword)
	}
}

func TransAddObject(w http.ResponseWriter, req any, sp *soap.SoapSessionInfo, devrun *devmap.DevRun) {
	logger.LogDebug("TransAddObjectResponse")

	if addInst, ok := req.(taskmodel.AddTask); ok {
		responseEnvelope := soap.NewAddObject(addInst.Name, sp.Env)
		TransmitXMLReq(responseEnvelope, w, sp.ContentType, devrun.AuthUsername, devrun.AuthPassword)
	}
}

func TransDeleteObject(w http.ResponseWriter, req any, sp *soap.SoapSessionInfo, devrun *devmap.DevRun) {
	logger.LogDebug("TransDeleteObjectResponse")

	if DelInst, ok := req.(string); ok {
		responseEnvelope := soap.NewDeleteObject(DelInst, sp.Env)
		TransmitXMLReq(responseEnvelope, w, sp.ContentType, devrun.AuthUsername, devrun.AuthPassword)
	}
}

func TransGetParameterNames(w http.ResponseWriter, req any, sp *soap.SoapSessionInfo, devrun *devmap.DevRun) {
	logger.LogDebug("TransGetParameterNames")

	if getlist, ok := req.(taskmodel.GetParamNamesTask); ok {
		responseEnvelope := soap.NewGetParameterNames(getlist, sp.Env)
		TransmitXMLReq(responseEnvelope, w, sp.ContentType, devrun.AuthUsername, devrun.AuthPassword)
	}
}

func TransGetParameterAttributes(w http.ResponseWriter, req any, sp *soap.SoapSessionInfo, devrun *devmap.DevRun) {
	logger.LogDebug("TransGetParameterAttributes")

	if getlist, ok := req.(taskmodel.GetParamAttrTask); ok {
		responseEnvelope := soap.NewGetParameterAttributes(getlist, sp.Env)
		TransmitXMLReq(responseEnvelope, w, sp.ContentType, devrun.AuthUsername, devrun.AuthPassword)
	}
}

func TransSetParameterAttributes(w http.ResponseWriter, req any, sp *soap.SoapSessionInfo, devrun *devmap.DevRun) {
	logger.LogDebug("TransSetParameterAttributes")

	if setlist, ok := req.([]taskmodel.SetParamAttrTask); ok {
		responseEnvelope := soap.NewSetParameterAttributes(setlist, sp.Env)
		TransmitXMLReq(responseEnvelope, w, sp.ContentType, devrun.AuthUsername, devrun.AuthPassword)
	}
}

func TransReboot(w http.ResponseWriter, req any, sp *soap.SoapSessionInfo, devrun *devmap.DevRun) {
	logger.LogDebug("TransReboot")

	responseEnvelope := soap.NewReboot(sp.Env)
	TransmitXMLReq(responseEnvelope, w, sp.ContentType, devrun.AuthUsername, devrun.AuthPassword)
}

func TransFactoryReset(w http.ResponseWriter, req any, sp *soap.SoapSessionInfo, devrun *devmap.DevRun) {
	logger.LogDebug("TransFactoryReset")

	responseEnvelope := soap.NewFactoryReset(sp.Env)
	TransmitXMLReq(responseEnvelope, w, sp.ContentType, devrun.AuthUsername, devrun.AuthPassword)
}

func TransDownload(w http.ResponseWriter, req any, sp *soap.SoapSessionInfo, devrun *devmap.DevRun) {
	logger.LogDebug("TransDownload")

	if download, ok := req.(taskmodel.DownloadTask); ok {
		responseEnvelope := soap.NewDownload(download, sp.Env)
		TransmitXMLReq(responseEnvelope, w, sp.ContentType, devrun.AuthUsername, devrun.AuthPassword)
	}
}

func TransUpload(w http.ResponseWriter, req any, sp *soap.SoapSessionInfo, devrun *devmap.DevRun) {
	logger.LogDebug("TransUpload", req, reflect.TypeOf(req))

	if upload, ok := req.(taskmodel.UploadTask); ok {
		responseEnvelope := soap.NewUpload(upload, sp.Env)
		TransmitXMLReq(responseEnvelope, w, sp.ContentType, devrun.AuthUsername, devrun.AuthPassword)
	}
}

func TransGetRPCMethods(w http.ResponseWriter, req any, sp *soap.SoapSessionInfo, devrun *devmap.DevRun) {
	logger.LogDebug("TransGetRPCMethods")

	responseEnvelope := soap.NewGetRPCMethods(sp.Env)
	TransmitXMLReq(responseEnvelope, w, sp.ContentType, devrun.AuthUsername, devrun.AuthPassword)
}

func TransTransferCompleteResponse(w http.ResponseWriter, req any, sp *soap.SoapSessionInfo, devrun *devmap.DevRun) {
	logger.LogDebug("TransTransferCompleteResponse")

	responseEnvelope := soap.NewTransferCompleteResponse(sp.Env)
	TransmitXMLReq(responseEnvelope, w, sp.ContentType, devrun.AuthUsername, devrun.AuthPassword)
}
