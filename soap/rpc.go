package soap

import (
	"context"
	"strconv"

	p "github.com/ecpartan/soap-server-tr069/internal/parsemap"
	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/soap-server-tr069/tasks"
)

func PrepareHeaderInfo(envelope any) EnvInfo {

	logger.LogDebug("PrepareHeaderInfo")

	envinfo := EnvInfo{}
	mp := p.GetXMLValueMap(envelope, "#attr")

	if mp != nil {
		soap_env_obj := p.GetXMLValue(mp, "xmlns:SOAP-ENV")
		soap_env := p.GetXMLValue(soap_env_obj, "#text").(string)
		if soap_env != "" {
			envinfo.SOAPENV = soap_env
		}

		soap_enc_obj := p.GetXMLValue(mp, "xmlns:SOAP-ENC")
		soap_enc := p.GetXMLValue(soap_enc_obj, "#text").(string)
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
	return envinfo
}

func NewInformResponse(env EnvInfo) *InformResponse {

	resp := &InformResponse{}

	resp.EnvInfo = env

	resp.Header.ID.MustUnderstand = "1"
	resp.Body.InformResponse.MaxEnvelopes = 1

	return resp
}

func NewGetParameterValues(paramlist tasks.GetParamTask, env EnvInfo) *GetParameterValues {
	logger.LogDebug("NewGetParameterValues")
	resp := &GetParameterValues{}

	resp.EnvInfo = mp.Env

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

func NewAddObject(obj string) *AddObject {
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

func NewDeleteObject(obj string) *DeleteObject {
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

func ParseEventCode(mp map[string]any) map[int]struct{} {
	codes := make(map[int]struct{})

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
