package devsoap

import (
	"fmt"
	"io"
	"net/http"

	"github.com/ecpartan/soap-server-tr069/httpserver"
	"github.com/ecpartan/soap-server-tr069/internal/apperror"
	"github.com/ecpartan/soap-server-tr069/internal/devmap"
	p "github.com/ecpartan/soap-server-tr069/internal/parsemap"
	logger "github.com/ecpartan/soap-server-tr069/log"
	repository "github.com/ecpartan/soap-server-tr069/repository/cache"
	"github.com/ecpartan/soap-server-tr069/server/handlers"
	"github.com/ecpartan/soap-server-tr069/soap"
	"github.com/ecpartan/soap-server-tr069/tasks"
	"github.com/julienschmidt/httprouter"
)

type handler struct {
	mapResponse *devmap.DevMap
	Cache       *repository.Cache
}

func NewHandler(mapResponse *devmap.DevMap, Cache *repository.Cache) handlers.Handler {
	return &handler{
		mapResponse: mapResponse,
		Cache:       Cache,
	}
}

func (h *handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodPost, "/", apperror.Middleware(h.PerformSoap))
}

func (h *handler) parseXML(addr string, mv map[string]any) soap.TaskResponseType {
	logger.LogDebug("ParseXML", mv)
	if mv == nil {
		return soap.ResponseUndefinded
	}
	envelope := p.GetXML(mv, "SOAP-ENV:Envelope")

	if envelope == nil {
		logger.LogDebug("envelope is not parseMapXML")
		return soap.ResponseUndefinded
	} else {

		mp := h.mapResponse.Get(addr)

		logger.LogDebug("mapresponse", mp)
		mp.Env = soap.PrepareHeaderInfo(envelope)

		xml_body := p.GetXML(envelope, "SOAP-ENV:Body")

		if xml_body == nil {
			return soap.ResponseUndefinded
		}
		var status = soap.ResponseUndefinded

		if ret, ok := p.GetXML(xml_body, "SOAP-ENV:Fault").(map[string]any); ok {
			status = soap.FaultResponse
			mp.Body = ret
		} else if ret, ok := p.GetXML(xml_body, "cwmp:Inform").(map[string]any); ok {
			logger.LogDebug("found Inform")
			if sn := p.GetXMLString(ret, "DeviceId.SerialNumber"); sn != "" {
				tasks.AddDevicetoTaskList(sn)

				paramlist := p.GetXML(ret, "ParameterList.ParameterValueStruct").([]any)
				logger.LogDebug("paramlist", paramlist)
				tasks.UpdateCacheBySerial(sn, paramlist, h.Cache, tasks.VALUES)

				mp.Body = ret
				mp.Serial = sn
				status = soap.Inform
			} else {
				return soap.ResponseUndefinded
			}
		} else if ret, ok := p.GetXML(xml_body, "cwmp:GetParameterValuesResponse").(map[string]any); ok {
			mp.Body = ret
			status = soap.GetParameterValuesResponse
		} else if ret, ok := p.GetXML(xml_body, "cwmp:SetParameterValuesResponse").(map[string]any); ok {
			mp.Body = ret
			status = soap.SetParameterValuesResponse
		} else if ret, ok := p.GetXML(xml_body, "cwmp:AddObjectResponse").(map[string]any); ok {
			mp.Body = ret
			status = soap.AddObjectResponse
		} else if ret, ok := p.GetXML(xml_body, "cwmp:DeleteObjectResponse").(map[string]any); ok {
			mp.Body = ret
			status = soap.DeleteObjectResponse
		} else if ret, ok := p.GetXML(xml_body, "cwmp:GetRPCMethodsResponse").(map[string]any); ok {
			mp.Body = ret
			status = soap.GetRPCMethodsResponse
		} else if ret, ok := p.GetXML(xml_body, "cwmp:GetParameterNamesResponse").(map[string]any); ok {
			mp.Body = ret
			status = soap.GetParameterNamesResponse
		} else if ret, ok := p.GetXML(xml_body, "cwmp:GetParameterAttributesResponse").(map[string]any); ok {
			mp.Body = ret
			status = soap.GetParameterAttributesResponse
		} else if ret, ok := p.GetXML(xml_body, "cwmp:SetParameterAttributesResponse").(map[string]any); ok {
			mp.Body = ret
			status = soap.SetParameterAttributesResponse
		} else if ret, ok := p.GetXML(xml_body, "cwmp:FactoryResetResponse").(map[string]any); ok {
			mp.Body = ret
			status = soap.FactoryResetResponse
		} else if ret, ok := p.GetXML(xml_body, "cwmp:UploadResponse").(map[string]any); ok {
			mp.Body = ret
			status = soap.UploadResponse
		} else if ret, ok := p.GetXML(xml_body, "cwmp:DownloadResponse").(map[string]any); ok {
			mp.Body = ret
			status = soap.DownloadResponse
		} else if ret, ok := p.GetXML(xml_body, "cwmp:TransferCompleteResponse").(map[string]any); ok {
			mp.Body = ret
			status = soap.TransferCompleteResponse
		} else {
			logger.LogDebug("not found")
			return soap.ResponseUndefinded
		}

		h.mapResponse.Set(addr, *mp)

		return status
	}
}

// Soap main handler
// @Summary Perform a SOAP request
// @Tags soap
// @Success 200
// @Router / [post]
func (h *handler) PerformSoap(w http.ResponseWriter, r *http.Request) error {
	soapRequestBytes, err := io.ReadAll(r.Body)

	if err != nil {
		return fmt.Errorf("could not read POST: %v", err)
	}

	addr := r.RemoteAddr

	logger.LogDebug("PerformSoap", string(soapRequestBytes))

	mv, err := p.ConvertXMLtoMap(soapRequestBytes)
	/*if err != nil {
		return fmt.Errorf("failed XML: %v", err)
	}*/

	logger.LogDebug("mv", err)

	mp := h.mapResponse.Get(addr)
	logger.LogDebug("mv", mp)

	if mv == nil {
		logger.LogDebug("End session")
		if tasks.GetTasks(w, addr, mp.ResponseTask, mp.SoapSessionInfo, h.mapResponse.Wg) {
			h.mapResponse.Delete(addr)
		}
		return nil
	}

	paramType := h.parseXML(addr, mv)

	logger.LogDebug("mapresponse", h.mapResponse)
	logger.LogDebug("found soap type", paramType, mp.Body)

	switch paramType {
	case soap.ResponseUndefinded:
		return fmt.Errorf("unknown XML Soap Type: %v", err)
	case soap.FaultResponse:
		tasks.ParseFaultResponse(mp)
	case soap.Inform:
		httpserver.TransInformResponse(w, mp.ResponseTask.Body, mp.SoapSessionInfo)
	case soap.GetParameterValuesResponse:
		tasks.ParseGetResponse(mp, h.Cache)
	case soap.SetParameterValuesResponse:
		tasks.ParseSetResponse(mp)
	case soap.AddObjectResponse:
		tasks.ParseAddResponse(mp)
	case soap.DeleteObjectResponse:
		tasks.ParseDeleteResponse(mp)
	case soap.GetRPCMethodsResponse:
		tasks.ParseGetRPCMethodsResponse(mp)
	case soap.GetParameterNamesResponse:
		tasks.ParseGetParameterNamesResponse(mp, h.Cache)
	case soap.GetParameterAttributesResponse:
		tasks.ParseGetParameterAttributesResponse(mp, h.Cache)
	default:

		break
	}
	if paramType != soap.Inform {
		tasks.GetTasks(w, addr, mp.ResponseTask, mp.SoapSessionInfo, h.mapResponse.Wg)
	}

	return nil
}
