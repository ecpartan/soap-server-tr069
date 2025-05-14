package soap

import (
	"context"
	"strconv"

	p "github.com/ecpartan/soap-server-tr069/internal/parsemap"
	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/soap-server-tr069/tasks"
)

func PrepareHeaderInfo(ctx context.Context, mp any) {

	logger.LogDebug("PrepareHeaderInfo")

	envinfo := EnvInfo{}

	if mp != nil {
		soap_env_obj := p.GetXMLValue(mp, "xmlns:SOAP-ENV")
		soap_env := p.GetXMLValue(soap_env_obj, "#text").(string)
		if soap_env != "" {
			envinfo.SOAPENV = soap_env
		}

		soap_enc_obj := p.GetXMLValue(mp, "xmlns:SOAP-ENC")
		soap_enc := p.GetXMLValue(soap_enc_obj, "#text").(string)
		logger.LogDebug("soap_env", soap_env)
		if soap_enc != "" {
			envinfo.SOAPENC = soap_enc
		}

		cwmp_obj := p.GetXMLValue(mp, "xmlns:cwmp")
		cwmp := p.GetXMLValue(cwmp_obj, "#text").(string)
		if cwmp != "" {
			envinfo.Cwmp = cwmp
		}

		xsi_obj := p.GetXMLValue(mp, "xmlns:xsi")
		xsi := p.GetXMLValue(xsi_obj, "#text").(string)
		if xsi != "" {
			envinfo.Xsi = xsi
		}

		xsd_obj := p.GetXMLValue(mp, "xmlns:xsd")
		xsd := p.GetXMLValue(xsd_obj, "#text").(string)
		if xsd != "" {
			envinfo.Xsd = xsd
		}
	} else {
		envinfo.SOAPENV = string(BNamespaceSoap12)
		envinfo.SOAPENC = string(BNamespaceEnc)
		envinfo.Cwmp = string(BNamespaceCwmp)
		envinfo.Xsd = string(BNamespaceXsd)
		envinfo.Xsi = string(BNamespaceXsi)
	}

	ctx = context.WithValue(ctx, "EnvInfo", envinfo)

}

func NewInformResponse(ctx context.Context, mp interface{}) *InformResponse {

	resp := &InformResponse{}
	envInfo, ok := ctx.Value("EnvInfo").(EnvInfo)
	if !ok {
		logger.LogDebug("NewInformResponse failed")
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

func NewGetParameterValues(ctx context.Context, paramlist tasks.GetParamTask) *GetParameterValues {
	logger.LogDebug("NewGetParameterValues")
	resp := &GetParameterValues{}

	envInfo, ok := ctx.Value("EnvInfo").(EnvInfo)
	if !ok {
		logger.LogDebug("NewGetParameterValues failed")
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

func NewSetParameterValues(ctx context.Context, paramlist []tasks.SetParamTask) *SetParameterValues {

	resp := &SetParameterValues{}

	envInfo, ok := ctx.Value("EnvInfo").(EnvInfo)
	if !ok {
		logger.LogDebug("NewSetParameterValues failed")
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

	logger.LogDebug("paramstruct: %v", paramstruct)
	resp.Body.SetParameterValues.ParameterList.ArrayType = "xsd:ParameterValueStruct[" + strconv.Itoa(len(paramlist)) + "]"

	resp.Header.ID.MustUnderstand = "1"

	return resp
}

func NewAddObject(ctx context.Context, obj string) *AddObject {
	resp := &AddObject{}

	envInfo, ok := ctx.Value("EnvInfo").(EnvInfo)
	if !ok {
		logger.LogDebug("NewAddObject failed")
		return nil
	}

	resp.SOAPENV = envInfo.SOAPENV
	resp.SOAPENC = envInfo.SOAPENC
	resp.Cwmp = envInfo.Cwmp
	resp.Xsd = envInfo.Xsd
	resp.Xsi = envInfo.Xsi

	resp.Body.AddObject.ObjectName = obj
	resp.Header.ID.MustUnderstand = "1"

	return resp
}

func NewDeleteObject(ctx context.Context, obj string) *DeleteObject {
	resp := &DeleteObject{}

	envInfo, ok := ctx.Value("EnvInfo").(EnvInfo)
	if !ok {
		logger.LogDebug("NewAddObject failed")
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

func CheckSoapType(ctx context.Context, addr string, mv map[string]interface{}) (map[string]interface{}, TaskType) {

	if mv == nil {
		return nil, ResponseEndSession
	}

	envelope := p.GetXMLValueS(mv, "SOAP-ENV:Envelope")

	if envelope == nil {
		logger.LogDebug("envelope is not parseMapXML")
		return nil, ResponseUndefinded
	} else {
		attrs := p.GetXMLValueMap(envelope, "#attr")

		PrepareHeaderInfo(ctx, attrs)

		bod := p.GetXMLValue(envelope, "SOAP-ENV:Body")

		if bod != nil {
			inf := p.GetXMLValue(bod, "cwmp:Inform")

			if inf != nil {
				logger.LogDebug("found Inform")
				serial := p.GetXMLValueS(inf, "DeviceId.SerialNumber.#text").(string)
				if serial != "" {
					tasks.AddDevicetoTaskList(serial, addr)
				}

				if _, ok := ctx.Value("DeviceID").(DeviceId); !ok {

					id := DeviceId{
						Manufacturer: p.GetXMLValueS(inf, "DeviceId.Manufacturer.#text").(string),
						OUI:          p.GetXMLValueS(inf, "DeviceId.OUI.#text").(string),
						ProductClass: p.GetXMLValueS(inf, "DeviceId.ProductClass.#text").(string),
						SerialNumber: p.GetXMLValueS(inf, "DeviceId.SerialNumber.#text").(string),
					}

					ctx = context.WithValue(ctx, "DeviceID", id)

				}

				return inf.(map[string]interface{}), Inform
			}

			ret := p.GetXMLValue(bod, "cwmp:GetParameterValuesResponse")

			if ret != nil {
				return ret.(map[string]interface{}), GetParameterValuesResponse
			}
			ret = p.GetXMLValue(bod, "cwmp:SetParameterValuesResponse")

			if ret != nil {
				return ret.(map[string]interface{}), SetParameterValuesResponse
			}

			ret = p.GetXMLValue(bod, "cwmp:AddObjectResponse")
			if ret != nil {
				return ret.(map[string]interface{}), AddObjectResponse
			}
			ret = p.GetXMLValue(bod, "cwmp:DeleteObjectResponse")
			if ret != nil {
				return ret.(map[string]interface{}), DeleteObjectResponse
			}
		}
	}
	return nil, ResponseUndefinded

}

type SetMap map[int]struct{}

func ParseEventCode(mp map[string]interface{}) SetMap {
	codes := make(SetMap)

	if mp != nil {
		events := p.GetXMLValueS(mp, "Event.EventStruct")
		if events == nil {
			return codes
		}

		logger.LogDebug("events", events)
		if list_events, ok := events.(map[string]any); ok {

			for event, map_event := range list_events {
				logger.LogDebug("event", map_event)
				if event == "EventCode" {

					eventCode := p.GetXMLValue(map_event, "#text").(string)

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
	return codes
}
