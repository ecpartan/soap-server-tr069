package soap

import (
	"strconv"

	p "github.com/ecpartan/soap-server-tr069/internal/parsemap"
	"github.com/ecpartan/soap-server-tr069/internal/taskmodel"
	logger "github.com/ecpartan/soap-server-tr069/log"
)

func PrepareHeaderInfo(envelope any) EnvInfo {

	logger.LogDebug("PrepareHeaderInfo")

	envinfo := EnvInfo{}

	if p.GetXML(envelope, "#attr") == nil {
		if soap_env, ok := p.GetXML(envelope, "#attr.xmlns:SOAP-ENV.#text").(string); ok {
			envinfo.SOAPENV = soap_env
		}

		if soap_enc, ok := p.GetXML(envelope, "#attr.xmlns:SOAP-ENC.#text").(string); ok {
			envinfo.SOAPENC = soap_enc
		}

		if cwmp, ok := p.GetXML(envelope, "#attr.xmlns:cwmp.#text").(string); ok {
			envinfo.Cwmp = cwmp
		}

		if xsi, ok := p.GetXML(envelope, "#attr.xmlns:xsi.#text").(string); ok {
			envinfo.Xsi = xsi
		}

		if xsd, ok := p.GetXML(envelope, "#attr.xmlns:xsd.#text").(string); ok {
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

func NewGetParameterValues(paramlist taskmodel.GetParamValTask, env EnvInfo) *GetParameterValues {
	logger.LogDebug("NewGetParameterValues")
	resp := &GetParameterValues{}

	resp.EnvInfo = env

	resp.Body.GetParameterValues.ParameterNames.String = paramlist.Name
	resp.Body.GetParameterValues.ParameterNames.ArrayType = "xsd:string[" + strconv.Itoa(len(paramlist.Name)) + "]"

	resp.Header.ID.MustUnderstand = "1"

	return resp
}

func NewSetParameterValues(paramlist []taskmodel.SetParamValTask, env EnvInfo) *SetParameterValues {

	resp := &SetParameterValues{}

	resp.EnvInfo = env

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

/*
	func NewSetParameterAttributes(paramlist []taskmodel.SetParamValTask, env EnvInfo) *SetParameterAttributes {
		resp := &SetParameterAttributes{}

		resp.EnvInfo = env

}
*/
func NewGetParameterAttributes(paramlist taskmodel.GetParamAttrTask, env EnvInfo) *GetParameterAttributes {
	resp := &GetParameterAttributes{}

	resp.EnvInfo = env
	/*for _, param := range paramlist {
		resp.Body.GetParameterAttributes.ParameterNames.String = append(resp.Body.GetParameterAttributes.ParameterNames.String, param.Name...)
	}*/
	resp.Body.GetParameterAttributes.ParameterNames.String = paramlist.Name
	resp.Body.GetParameterAttributes.ParameterNames.ArrayType = "xsd:string[" + strconv.Itoa(len(paramlist.Name)) + "]"

	resp.Header.ID.MustUnderstand = "1"
	logger.LogDebug("paramstruct: %v", resp)

	return resp
}
func NewGetParameterNames(paramlist taskmodel.GetParamNamesTask, env EnvInfo) *GetParameterNames {

	resp := &GetParameterNames{}

	resp.EnvInfo = env

	resp.Body.GetParameterNames.ParameterPath = paramlist.ParameterPath
	resp.Body.GetParameterNames.NextLevel = strconv.Itoa(paramlist.NextLevel)

	logger.LogDebug("paramstruct: %v", resp)

	resp.Header.ID.MustUnderstand = "1"

	return resp
}

func NewGetRPCMethods(paramlist taskmodel.GetParamNamesTask, env EnvInfo) *GetRPCMethods {

	resp := &GetRPCMethods{}

	resp.EnvInfo = env

	logger.LogDebug("paramstruct: %v", resp)

	resp.Header.ID.MustUnderstand = "1"

	return resp
}

func NewAddObject(obj string, env EnvInfo) *AddObject {
	resp := &AddObject{}

	resp.EnvInfo = env

	resp.Body.AddObject.ObjectName = obj
	resp.Header.ID.MustUnderstand = "1"

	return resp
}

func NewDeleteObject(obj string, env EnvInfo) *DeleteObject {
	resp := &DeleteObject{}
	resp.EnvInfo = env

	resp.Body.DeleteObject.ObjectName = obj
	resp.Header.ID.MustUnderstand = "1"

	return resp
}

func ParseEventCode(mp map[string]any) map[int]struct{} {
	codes := make(map[int]struct{})

	if mp == nil {
		return codes
	}

	events := p.GetXML(mp, "Event.EventStruct")
	if events == nil {
		return codes
	}

	logger.LogDebug("events", events)
	if list_events, ok := events.(map[string]any); ok {

		for event, map_event := range list_events {
			logger.LogDebug("event", map_event)
			if event == "EventCode" {

				eventCode := p.GetXML(map_event, "#text").(string)

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

	return codes
}
