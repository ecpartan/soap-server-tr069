package tasks

import (
	"github.com/ecpartan/soap-server-tr069/internal/devmap"
	"github.com/ecpartan/soap-server-tr069/internal/devmodel"
	logger "github.com/ecpartan/soap-server-tr069/log"
	p "github.com/ecpartan/soap-server-tr069/pkg/parsemap"
	repository "github.com/ecpartan/soap-server-tr069/repository/cache"
	"github.com/ecpartan/soap-server-tr069/utils/semaphore"
)

func ParseResponseDef(mp *devmap.DevID, rpc string) {
	logger.LogDebug(rpc)
	logger.LogDebug("body, ", mp.Body)

	if mp.RespChan == nil {
		mp.RespChan <- devmodel.SoapResponse{Code: "-1", Method: rpc}
		return
	}
	go func() {
		logger.LogDebug(rpc, mp.WaitEvent)

		if mp.WaitEvent > -1 {
			semaphore.Acquire(mp.Serial)
		} else {
			mp.RespChan <- devmodel.SoapResponse{Code: "0", Method: rpc}
		}
	}()
}

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
		go func() {
			logger.LogDebug("PSR", mp.WaitEvent)

			if mp.WaitEvent > -1 {
				semaphore.Acquire(mp.Serial)
			} else {
				mp.RespChan <- devmodel.SoapResponse{Code: status, Method: "Set"}
			}
		}()
		return
	}

	mp.RespChan <- devmodel.SoapResponse{Code: "-1", Method: "Set"}
}

func CheckWaitEvent(mp *devmap.DevID) {
	logger.LogDebug("CheckWaitEvent", mp.EventCodes, mp.WaitEvent)

	if mp.WaitEvent > -1 {
		if _, ok := mp.EventCodes[mp.WaitEvent]; ok {

			err := semaphore.Release(mp.Serial)
			if err != nil {
				logger.LogErr("CheckWaitEvent", err)
			}
			mp.WaitEvent = -1
			mp.RespChan <- devmodel.SoapResponse{Code: "0", Method: "Get"}
		}
	}
}
func ParseGetResponse(mp *devmap.DevID, l *repository.Cache) {

	logger.LogDebug("ParseGetResponse")
	logger.LogDebug("body,", mp.Body)

	if paramlist, ok := p.GetXML(mp.Body, "ParameterList.ParameterValueStruct").([]any); !ok {
		paramlist = make([]any, 0)
		if params, ok := p.GetXML(mp.Body, "ParameterList.ParameterValueStruct").(map[string]any); ok {
			paramlist = append(paramlist, params)
		}
		if paramlist == nil || mp.RespChan == nil {
			mp.RespChan <- devmodel.SoapResponse{Code: "-1", Method: "Get"}
			return
		}
		repository.UpdateCacheBySerial(mp.Serial, paramlist, l, repository.VALUES)
	}
	go func() {
		logger.LogDebug("PGR", mp.WaitEvent)

		if mp.WaitEvent > -1 {
			semaphore.Acquire(mp.Serial)
		} else {
			mp.RespChan <- devmodel.SoapResponse{Code: "0", Method: "Get"}
		}
	}()
}

func ParseGetRPCMethodsResponse(mp *devmap.DevID) {
	ParseResponseDef(mp, "GetRPCMethods")
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

	go func() {
		logger.LogDebug("GetParameterAttributes", mp.WaitEvent)
		if mp.WaitEvent > -1 {
			semaphore.Acquire(mp.Serial)
		} else {
			mp.RespChan <- devmodel.SoapResponse{Code: "0", Method: "GetParameterAttributes"}
		}
	}()

}

func ParseSetParameterAttributesResponse(mp *devmap.DevID, l *repository.Cache) {
	ParseResponseDef(mp, "SetParameterAttributes")
}

func ParseDownloadResponse(mp *devmap.DevID) {
	ParseResponseDef(mp, "Download")
}

func ParseUploadResponse(mp *devmap.DevID) {
	ParseResponseDef(mp, "Upload")
}
func ParseRebootResponse(mp *devmap.DevID) {
	ParseResponseDef(mp, "Reboot")
}
func ParseFactoryResetResponse(mp *devmap.DevID) {
	ParseResponseDef(mp, "FactoryReset")
}

func ParseTransferCompleteResponse(mp *devmap.DevID) {
	ParseResponseDef(mp, "TransferCompleteResponse")
}
