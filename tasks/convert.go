package tasks

import (
	"strings"

	"github.com/dgrijalva/lfu-go"
	p "github.com/ecpartan/soap-server-tr069/internal/parsemap"
	logger "github.com/ecpartan/soap-server-tr069/log"
)

func SetValueInJSON(iface any, path string, value any) any {
	m := iface.(map[string]any)
	split := strings.Split(path, ".")
	for k, v := range m {
		if strings.EqualFold(k, split[0]) {
			if len(split) == 1 {
				m[k] = value
				return m
			}
			switch v.(type) {
			case map[string]any:
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
		newMap := make(map[string]any)
		newMap[split[len(split)-1]] = value
		for i := len(split) - 2; i > 0; i-- {
			mTmp := make(map[string]any)
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

func ParseAddResponse(xml_body any, respchan chan any) {

	logger.LogDebug("ParseAddResponse")
	logger.LogDebug("body,", xml_body)

	if status, ok := p.GetXMLValueS(xml_body, "Status.#text").(string); ok {
		if status == "1" || status == "0" {
			logger.LogDebug("Return:", status)

			if number, ok := p.GetXMLValueS(xml_body, "InstanceNumber.#text").(string); ok {
				respchan <- number
				close(respchan)
			}
		}
	}
}

func ParseDeleteResponse(xml_body any, respchan chan<- any) {

	logger.LogDebug("ParseAddResponse")
	logger.LogDebug("body,", xml_body)

	if status, ok := p.GetXMLValue(xml_body, "Status").(string); ok {
		if status == "1" || status == "0" {
			respchan <- status
			close(respchan)
		}
	}
}

func ParseSetResponse(xml_body any, respchan chan<- any) {

	logger.LogDebug("ParseAddResponse")
	logger.LogDebug("body,", xml_body)

	if respchan == nil {
		return
	}
	if status, ok := p.GetXMLValue(xml_body, "Status").(string); ok {
		if status == "1" || status == "0" {
			respchan <- status
			close(respchan)
		}
	}
}

func UpdateCacheBySerial(serial string, paramlist []any, l *lfu.Cache) {

	device_cache := l.Get(serial)
	if device_cache == nil {
		device_cache = make(map[string]any)
	}
	if device_cache, ok := device_cache.(map[string]any); ok {
		logger.LogDebug("device_cache", device_cache)
		new_device_cache := updateJsonParams(paramlist, device_cache)
		l.Set(serial, new_device_cache)
	}
}
func ParseGetResponse(xml_body any, serial string, respchan chan any, l *lfu.Cache) {

	logger.LogDebug("ParseGetResponse")
	logger.LogDebug("body,", xml_body)

	paramlist := p.GetXMLValueS(xml_body, "ParameterList.ParameterValueStruct").([]any)

	if paramlist == nil || respchan == nil {
		close(respchan)
		return
	}
	UpdateCacheBySerial(serial, paramlist, l)

	respchan <- paramlist
	close(respchan)
}
