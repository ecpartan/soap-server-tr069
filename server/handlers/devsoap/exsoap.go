package devsoap

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ecpartan/soap-server-tr069/httpserver"
	"github.com/ecpartan/soap-server-tr069/internal/apperror"
	"github.com/ecpartan/soap-server-tr069/internal/devmap"
	p "github.com/ecpartan/soap-server-tr069/internal/parsemap"
	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/soap-server-tr069/pkg/httplogger"
	"github.com/ecpartan/soap-server-tr069/pkg/jrpc2/methods/response"
	"github.com/ecpartan/soap-server-tr069/pkg/monitoring"
	repository "github.com/ecpartan/soap-server-tr069/repository/cache"
	"github.com/ecpartan/soap-server-tr069/repository/db/domain/entity"
	usecase_device "github.com/ecpartan/soap-server-tr069/repository/db/domain/usecase/device"
	"github.com/ecpartan/soap-server-tr069/server/handlers"
	"github.com/ecpartan/soap-server-tr069/soap"
	"github.com/ecpartan/soap-server-tr069/tasks"
	"github.com/ecpartan/soap-server-tr069/tasks/tasker"
	"github.com/julienschmidt/httprouter"
)

type handler struct {
	mapResponse *devmap.DevMap
	Cache       *repository.Cache
	service     *usecase_device.Service
	execTasks   *tasker.Tasker
	metrics     *monitoring.MetricsService
}

func NewHandler(mapResponse *devmap.DevMap, Cache *repository.Cache, service *usecase_device.Service, execTasks *tasker.Tasker, metrics *monitoring.MetricsService) handlers.Handler {
	return &handler{
		mapResponse: mapResponse,
		Cache:       Cache,
		service:     service,
		execTasks:   execTasks,
		metrics:     metrics,
	}
}

func (h *handler) Register(router *httprouter.Router) {
	/*err := os.Remove("./output")

	file, err := os.OpenFile("./output", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	httplg := httplog.LoggerWithConfig(httplog.LoggerConfig{
		RouterName:  "FillBodyFormatter",
		Formatter:   httplog.FullFormatterWithRequestAndResponseHeadersAndBody,
		CaptureBody: true,
		Output:      file,
	})

	router.Handler(http.MethodPost, "/", httplg(http.HandlerFunc(h.PerformSoap)))*/

	router.Handler(http.MethodPost, "/", httplogger.NewHTTPLogger().Middleware(apperror.Middleware(h.PerformSoap)))
	//router.Handler(http.MethodPost, "/", LoggingMiddleware(h.PerformSoap))

}

func (h *handler) updateDeviceDB(serial string, cache *repository.Cache, inform map[string]any) error {
	dev, err := h.service.GetOneBySn(serial)

	logger.LogDebug("GetDevicebySerial", dev, err)
	if err != nil {
		return err
	}

	var manufacturer, model, oui, sw, hw, datamodel, crurl string
	devid := p.GetXML(inform, "DeviceId")
	manufacturer = p.GetXMLString(devid, "Manufacturer")
	oui = p.GetXMLString(inform, "OUI")

	tree := cache.Get(serial)
	if p.GetXML(tree, soap.Pref98) != nil {
		datamodel = "98"
	} else {
		datamodel = "181"
	}
	crurl = p.GetXMLValue(tree, soap.CR_URL)
	hw = p.GetXMLValue(tree, soap.HW_V)
	sw = p.GetXMLValue(tree, soap.SW_V)
	model = p.GetXMLValue(tree, soap.MODELNAME)

	if dev == nil {
		view := entity.NewDeviceView(serial, manufacturer, model, oui, sw, hw, datamodel, crurl)
		_, err := h.service.CreateDevice(*view)
		return err
	} else {
		dev.Datamodel = datamodel
		dev.CrURL = crurl
		dev.HwVersion = hw
		dev.SwVersion = sw
		dev.Model = model
		dev.OUI = oui
		dev.Manufacturer = manufacturer

		return h.service.UpdateDevice(dev)
	}
}

func (h *handler) parseXML(r *http.Request, addr string, mv map[string]any) soap.TaskResponseType {
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
				if authview, err := h.service.GetAuthDeviceBySn(sn); err == nil {
					logger.LogDebug("authview", authview)
					mp.AuthUsername = authview.Username
					mp.AuthPassword = authview.Password
					if !check_authorization(r, mp) {
						return soap.ResponseUnauthorized
					}
				} else {
					logger.LogDebug("authview", err)
				}

				h.execTasks.ExecTasks.AddDevicetoTaskList(sn)
				h.metrics.AddDevInst()
				paramlist := p.GetXML(ret, "ParameterList.ParameterValueStruct").([]any)
				logger.LogDebug("paramlist", paramlist)
				tasks.UpdateCacheBySerial(sn, paramlist, h.Cache, tasks.VALUES)
				h.updateDeviceDB(sn, h.Cache, ret)

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
		} else if ret, ok := p.GetXML(xml_body, "cwmp:TransferComplete").(map[string]any); ok {
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

func check_authorization(r *http.Request, mp *devmap.DevID) bool {
	requiredUser := mp.AuthUsername
	requiredPassword := mp.AuthPassword

	if requiredUser == "" || requiredPassword == "" {
		return true
	}

	user, password, hasAuth := r.BasicAuth()
	logger.LogDebug("Basic auth: ", hasAuth, user, password)
	if hasAuth && user == requiredUser && password == requiredPassword {
		return true
	}

	return false
}

func (h *handler) performSoap(w http.ResponseWriter, r *http.Request) (error, time.Time) {

	t := h.metrics.AddActiveInst()

	soapRequestBytes, err := io.ReadAll(r.Body)

	if err != nil {
		return fmt.Errorf("could not read POST: %v", err), t
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
		if tasks.GetTasks(w, addr, mp.ResponseTask, mp.SoapSessionInfo, h.mapResponse.Wg, h.execTasks.ExecTasks) {
			response.CloseChannelBySn(mp.Serial)
			h.mapResponse.Delete(addr)
		}

		return nil, t
	}

	paramType := h.parseXML(r, addr, mv)

	if paramType == soap.ResponseUnauthorized {
		return apperror.ErrUnAuthorized, t
	}

	if err != nil {
		return err, t
	}

	logger.LogDebug("mapresponse", h.mapResponse)
	logger.LogDebug("found soap type", paramType, mp.Body)

	switch paramType {
	case soap.ResponseUndefinded:
		return nil, t //fmt.Errorf("unknown XML Soap Type: %v", err), t
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
	case soap.SetParameterAttributesResponse:
		tasks.ParseSetParameterAttributesResponse(mp, h.Cache)
	case soap.DownloadResponse:
		tasks.ParseDownloadResponse(mp)
	case soap.UploadResponse:
		tasks.ParseUploadResponse(mp)
	case soap.RebootResponse:
		tasks.ParseRebootResponse(mp)
	case soap.FactoryResetResponse:
		tasks.ParseFactoryResetResponse(mp)
	case soap.TransferCompleteResponse:
		httpserver.TransTransferCompleteResponse(w, mp.ResponseTask.Body, mp.SoapSessionInfo)
		//tasks.ParseTransferCompleteResponse(mp)
	default:

		break
	}
	if paramType != soap.Inform {
		tasks.GetTasks(w, addr, mp.ResponseTask, mp.SoapSessionInfo, h.mapResponse.Wg, h.execTasks.ExecTasks)
	}

	return nil, t
}

// Soap main handler
// @Summary Perform a SOAP request
// @Tags soap
// @Success 200
// @Router / [post]
func (h *handler) PerformSoap(w http.ResponseWriter, r *http.Request) error {

	err, t := h.performSoap(w, r)

	if err != nil {
		logger.LogErr("PerformSoap", err)
		h.metrics.MetricsResultReq(t, err.Error())
		return err
	}

	h.metrics.MetricsResultReq(t, "OK")

	return nil
}
