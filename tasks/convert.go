package tasks

import (
	"strings"

	p "github.com/ecpartan/soap-server-tr069/internal/parsemap"
	logger "github.com/ecpartan/soap-server-tr069/log"
)

func SetValueInJSON(iface interface{}, path string, value interface{}) interface{} {
	m := iface.(map[string]interface{})
	split := strings.Split(path, ".")
	for k, v := range m {
		if strings.EqualFold(k, split[0]) {
			if len(split) == 1 {
				m[k] = value
				return m
			}
			switch v.(type) {
			case map[string]interface{}:
				return SetValueInJSON(v, strings.Join(split[1:], "."), value)
			default:
				return m
			}
		}
	}
	// path not found -> create
	if len(split) == 1 {
		m[split[0]] = value
	} else {
		newMap := make(map[string]interface{})
		newMap[split[len(split)-1]] = value
		for i := len(split) - 2; i > 0; i-- {
			mTmp := make(map[string]interface{})
			mTmp[split[i]] = newMap
			newMap = mTmp
		}
		m[split[0]] = newMap
	}
	return m
}

/*
	func SetBodyToMap(paramlist []any) map[string]any {
		mp := make((map[string]any), len(paramlist))

		for _, v := range paramlist {
			name := p.GetXMLValue(p.GetXMLValue(v, "Name"), "#text")
			value := p.GetXMLValue(p.GetXMLValue(v, "Value"), "#text")

			if n, ok := name.(string); ok {
				if val, ok := value.(string); ok {
					splitstr := strings.Split(n, ".")
					prev := splitstr[0]
					curr_mp := mp
					for i := 1; i < len(splitstr); i++ {
						if _, ok := curr_mp[prev]; !ok {
							curr_mp[prev] = make(map[string]any)
							curr_mp = curr_mp[prev].(map[string]any)
							prev = splitstr[i]
						} else {
							curr_mp = curr_mp[prev].(map[string]any)
							prev = splitstr[i]
						}
					}
					curr_mp[prev] = val
				}
			}
		}
		return mp
	}
*/
func updateJsonParams(paramlist []any, mp map[string]any) map[string]any {

	for _, v := range paramlist {
		name := p.GetXMLValueS(v, "Name.#text")
		value := p.GetXMLValueS(v, "Value.#text")
		if n, ok := name.(string); ok {
			if val, ok := value.(string); ok {
				SetValueInJSON(mp, n, val)
			}
		}
	}

	logger.LogDebug("updateJsonParams", mp)
	return mp
}

func ParseAddResponse(xml_body any, host string) {

	logger.LogDebug("ParseAddResponse")
	logger.LogDebug("body,", xml_body)
	resp := s.mapResponse[host]
	if status, ok := p.GetXMLValueS(xml_body, "Status.#text").(string); ok {
		s.log(status)
		if status == "1" || status == "0" {
			s.log("Return:", status)

			if number, ok := p.GetXMLValueS(xml_body, "InstanceNumber.#text").(string); ok {

				resp.respChan <- number

				//close(resp.respChan)
				return
			}
		}
	}

	//resp.respChan <- struct{}{}
	//close(resp.respChan)
}

func ParseDeleteResponse(xml_body any, host string) {

	logger.LogDebug("ParseDeleteResponse")
	s.log(xml_body)
	resp := s.mapResponse[host]

	if status, ok := GetXMLValue(xml_body, "Status").(string); ok {
		if status == "1" || status == "0" {
			resp.respChan <- status
			close(resp.respChan)
			return
		}
	}

	//resp.respChan <- struct{}{}
	//close(resp.respChan)
}

func ParseSetResponse(xml_body any, host string) {

	s.log("ParseSetResponse")
	s.log(xml_body)
	resp := s.mapResponse[host]
	if resp.respChan == nil {
		return
	}
	if status, ok := p.GetXMLValue(xml_body, "Status").(string); ok {
		if status == "1" || status == "0" {
			resp.respChan <- status
			close(resp.respChan)
			return
		}
	}
}

func ParseGetResponse(xml_body any, host string) {

	s.log("ParseGetResponse")

	paramlist := GetXMLValueS(xml_body, "ParameterList.ParameterValueStruct").([]any)
	if paramlist == nil {
		return
	}
	resp := s.mapResponse[host]

	if deviceID, ok := s.context.Value("DeviceID").(DeviceId); ok {
		device_cache := s.Cache.Get(deviceID.SerialNumber)
		if device_cache == nil {
			device_cache = make(map[string]any)
		}
		if device_cache, ok := device_cache.(map[string]any); ok {
			s.log("device_cache", device_cache)
			new_device_cache := updateJsonParams(paramlist, device_cache)
			s.Cache.Set(deviceID.SerialNumber, new_device_cache)
		}
	}

	if resp.respChan == nil {
		return
	}
	resp.respChan <- paramlist
	//close(resp.respChan)
}
