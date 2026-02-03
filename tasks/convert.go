package tasks

import (
	"github.com/ecpartan/soap-server-tr069/internal/devmap"
	"github.com/ecpartan/soap-server-tr069/internal/devmodel"
	p "github.com/ecpartan/soap-server-tr069/internal/parsemap"
	logger "github.com/ecpartan/soap-server-tr069/log"
	repository "github.com/ecpartan/soap-server-tr069/repository/cache"
)

func ParseFaultResponse(mp *devmap.DevID) {
	logger.LogDebug("ParseFaultResponse")
	logger.LogDebug("body,", mp.Body)

	faultcode := p.GetXMLString(mp.Body, "faultcode")
	faultstring := p.GetXMLString(mp.Body, "faultstring")

	if faultcode != "" && faultstring != "" {
		ret := devmodel.SoapResponse{Code: faultcode + " " + faultstring, Num: "404", Method: "Fault"}
		mp.RespChan <- ret
		return
	}
	mp.RespChan <- devmodel.SoapResponse{Code: "0", Method: "Fault"}
}

func ParseAddResponse(mp *devmap.DevID) {

	logger.LogDebug("ParseAddResponse")
	logger.LogDebug("body,", mp.Body)

	if status := p.GetXMLString(mp.Body, "Status"); status == "1" || status == "0" {
		logger.LogDebug("Return:", status)
		if number := p.GetXMLString(mp.Body, "InstanceNumber"); number != "" {
			ret := devmodel.SoapResponse{Code: status, Num: number, Method: "Add"}
			mp.RespChan <- ret
			return
		}
	}

	mp.RespChan <- devmodel.SoapResponse{Code: "-1", Method: "Add"}
}

func ParseDeleteResponse(mp *devmap.DevID) {

	logger.LogDebug("ParseDeleteResponse")
	logger.LogDebug("body,", mp.Body)

	if status := p.GetXMLString(mp.Body, "Status"); status == "1" || status == "0" {
		mp.RespChan <- devmodel.SoapResponse{Code: status, Method: "Delete"}
		return
	}

	mp.RespChan <- devmodel.SoapResponse{Code: "-1", Method: "Delete"}
}

func ParseSetResponse(mp *devmap.DevID) {

	logger.LogDebug("ParseSetResponse")
	logger.LogDebug("body,", mp.Body)

	if mp.RespChan == nil {
		return
	}

	if status := p.GetXMLString(mp.Body, "Status"); status == "1" || status == "0" {
		mp.RespChan <- devmodel.SoapResponse{Code: status, Method: "Set"}
		return
	}

	mp.RespChan <- devmodel.SoapResponse{Code: "-1", Method: "Set"}
}

func ParseGetResponse(mp *devmap.DevID, l *repository.Cache) {

	logger.LogDebug("ParseGetResponse")
	logger.LogDebug("body,", mp.Body)

	paramlist := p.GetXML(mp.Body, "ParameterList.ParameterValueStruct").([]any)

	if paramlist == nil || mp.RespChan == nil {
		mp.RespChan <- devmodel.SoapResponse{Code: "-1", Method: "Get"}
		return
	}
	repository.UpdateCacheBySerial(mp.Serial, paramlist, l, repository.VALUES)

	mp.RespChan <- devmodel.SoapResponse{Code: "0", Method: "Get"}
}

func ParseGetRPCMethodsResponse(mp *devmap.DevID) {
	logger.LogDebug("ParseGetRPCMethodsResponse")
	logger.LogDebug("body,", mp.Body)

	if mp.RespChan == nil {
		mp.RespChan <- devmodel.SoapResponse{Code: "-1", Method: "GetRPCMethods"}
		return
	}

	mp.RespChan <- devmodel.SoapResponse{Code: "0", Method: "GetRPCMethods"}
}

func ParseGetParameterNamesResponse(mp *devmap.DevID, l *repository.Cache) {
	logger.LogDebug("ParseGetParameterNamesResponse")
	logger.LogDebug("body,", mp.Body)

	if mp.RespChan == nil {
		mp.RespChan <- devmodel.SoapResponse{Code: "-1", Method: "GetParameterNames"}
		return
	}
	paramlist := p.GetXML(mp.Body, "ParameterList.ParameterInfoStruct").([]any)

	repository.UpdateCacheBySerial(mp.Serial, paramlist, l, repository.NAMES)

	mp.RespChan <- devmodel.SoapResponse{Code: "0", Method: "GetParameterNames"}
}

func ParseGetParameterAttributesResponse(mp *devmap.DevID, l *repository.Cache) {
	logger.LogDebug("ParseGetParameterAttributesResponse")
	logger.LogDebug("body,", mp.Body)

	if mp.RespChan == nil {
		mp.RespChan <- devmodel.SoapResponse{Code: "-1", Method: "GetParameterAttributes"}
		return
	}

	paramlist := p.GetXML(mp.Body, "ParameterList.ParameterAttributeStruct").([]any)

	repository.UpdateCacheBySerial(mp.Serial, paramlist, l, repository.ATTRS)

	mp.RespChan <- devmodel.SoapResponse{Code: "0", Method: "GetParameterAttributes"}
}

func ParseSetParameterAttributesResponse(mp *devmap.DevID, l *repository.Cache) {
	logger.LogDebug("ParseSetParameterAttributesResponse")
	logger.LogDebug("body,", mp.Body)

	if mp.RespChan == nil {
		mp.RespChan <- devmodel.SoapResponse{Code: "-1", Method: "SetParameterAttributes"}
		return
	}

	//paramlist := p.GetXML(mp.Body, "ParameterList.ParameterAttributeStruct").([]any)

	//UpdateCacheBySerial(mp.Serial, paramlist, l, ATTRS)

	mp.RespChan <- devmodel.SoapResponse{Code: "0", Method: "SetParameterAttributes"}
}

func ParseDownloadResponse(mp *devmap.DevID) {
	logger.LogDebug("ParseDownloadResponse")
	logger.LogDebug("body, ", mp.Body)

	if mp.RespChan == nil {
		mp.RespChan <- devmodel.SoapResponse{Code: "-1", Method: "Download"}
		return
	}

	mp.RespChan <- devmodel.SoapResponse{Code: "0", Method: "Download"}
}

func ParseUploadResponse(mp *devmap.DevID) {
	logger.LogDebug("ParseUploadResponse")
	logger.LogDebug("body, ", mp.Body)

	if mp.RespChan == nil {
		mp.RespChan <- devmodel.SoapResponse{Code: "-1", Method: "Upload"}
		return
	}

	mp.RespChan <- devmodel.SoapResponse{Code: "0", Method: "Upload"}
}

func ParseRebootResponse(mp *devmap.DevID) {
	logger.LogDebug("ParseRebootResponse")
	logger.LogDebug("body, ", mp.Body)

	if mp.RespChan == nil {
		mp.RespChan <- devmodel.SoapResponse{Code: "-1", Method: "Reboot"}
		return
	}

	mp.RespChan <- devmodel.SoapResponse{Code: "0", Method: "Reboot"}
}

func ParseFactoryResetResponse(mp *devmap.DevID) {
	logger.LogDebug("ParseFactoryResetResponse")
	logger.LogDebug("body, ", mp.Body)

	if mp.RespChan == nil {
		mp.RespChan <- devmodel.SoapResponse{Code: "-1", Method: "FactoryReset"}
		return
	}

	mp.RespChan <- devmodel.SoapResponse{Code: "0", Method: "FactoryReset"}
}

func ParseTransferCompleteResponse(mp *devmap.DevID) {
	logger.LogDebug("ParseTransferCompleteResponse")
	logger.LogDebug("body, ", mp.Body)

	if mp.RespChan == nil {
		mp.RespChan <- devmodel.SoapResponse{Code: "-1", Method: "TransferComplete"}
		return
	}

	mp.RespChan <- devmodel.SoapResponse{Code: "0", Method: "TransferComplete"}
}
