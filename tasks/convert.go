package tasks

import (
	"reflect"
	"strings"

	"github.com/ecpartan/soap-server-tr069/internal/devmap"
	"github.com/ecpartan/soap-server-tr069/internal/devmodel"
	p "github.com/ecpartan/soap-server-tr069/internal/parsemap"
	logger "github.com/ecpartan/soap-server-tr069/log"
	repository "github.com/ecpartan/soap-server-tr069/repository/cache"
)

type NodeStore struct {
	Value    string
	Type     string
	Writable string
	Notify   string
}

func setEndValue(m map[string]any, key string, value any) any {
	if v, ok := value.(NodeStore); !ok {
		if _, ok := m[key]; !ok {
			m[key] = map[string]any{"Value": v}
			logger.LogDebug("SetValueInJSON", "Value new", m[key])

			return m
		}
		if mp, ok := m[key].(map[string]any); ok {
			if _, ok := mp["Value"]; !ok {
				mp["Value"] = v
				logger.LogDebug("SetValueInJSON", "Value already exist111", mp["Value"])
				return m
			}
			if nodeVal, ok := mp["Value"].(NodeStore); ok {
				logger.LogDebug("SetValueInJSON", "Value already exist", nodeVal)
				if v.Value != "" {
					nodeVal.Value = v.Value
				}
				if v.Writable != "" {
					nodeVal.Writable = v.Writable
				}
				if v.Notify != "" {
					nodeVal.Notify = v.Notify
				}
				if v.Type != "" {
					nodeVal.Type = v.Type
				}
				mp["Value"] = nodeVal
				return m
			}
		}
	}
	return m
}

func SetValueInJSON(iface any, path string, key, value string) any {
	m := iface.(map[string]any)
	split := strings.Split(path, ".")
	for k, v := range m {
		if strings.EqualFold(k, split[0]) {
			if len(split) == 1 {
				if _, ok := m[k]; !ok {
					m[k] = map[string]any{key: value}
					return m
				} else {
					logger.LogDebug("mp", reflect.TypeOf(m[k]), path)
					if mp, ok := m[k].(map[string]any); ok {
						logger.LogDebug("mp", mp)

						mp[key] = value
						m[k] = mp

						return m
					} else {
						m[k] = map[string]any{key: value}
						return m
					}
				}
			}
			switch v.(type) {
			case map[string]any:
				return SetValueInJSON(v, strings.Join(split[1:], "."), key, value)
			default:
				return m
			}
		}
	}
	// path not found -> create
	if len(split) == 1 {
		logger.LogDebug("mp", m[split[0]], reflect.TypeOf(m[split[0]]))
		if m[split[0]] == nil {
			m[split[0]] = map[string]any{key: value}
		} else {
			if mp, ok := m[split[0]].(map[string]any); ok {
				logger.LogDebug("mp", mp)

				mp[key] = value
				m[split[0]] = mp
			} else {
				m[split[0]] = map[string]any{key: value}

			}
		}
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

func updateDeviceValues(paramlist []any, mp map[string]any) map[string]any {

	for _, v := range paramlist {
		name := p.GetXML(v, "Name.#text")
		value := p.GetXML(v, "Value.#text")
		xtype := p.GetXML(v, "Value.#attr.xsi:type.#text")
		if n, ok := name.(string); ok {
			if val, ok := value.(string); ok {
				if t, ok := xtype.(string); ok {
					SetValueInJSON(mp, n, "Value", val)
					SetValueInJSON(mp, n, "Type", t)
				}
			}
		}
	}

	logger.LogDebug("updateDeviceValues", mp)
	return mp
}

func updateDeviceAttrs(paramlist []any, mp map[string]any) map[string]any {

	for _, v := range paramlist {
		name := p.GetXML(v, "Name.#text")
		notify := p.GetXML(v, "Notification.#text")
		if n, ok := name.(string); ok {
			if not, ok := notify.(string); ok {
				SetValueInJSON(mp, n, "Notify", not)
			}
		}
	}

	logger.LogDebug("updateDeviceAttrs", mp)
	return mp
}

func updateDeviceNames(paramlist []any, mp map[string]any) map[string]any {
	logger.LogDebug("updateDeviceNames", mp)

	for _, v := range paramlist {
		name := p.GetXML(v, "Name.#text")
		writable := p.GetXML(v, "Writable.#text")
		if n, ok := name.(string); ok {
			if w, ok := writable.(string); ok {
				SetValueInJSON(mp, n, "Writable", w)
			}
		}
	}

	logger.LogDebug("updateDeviceNames", mp)
	return mp
}

type ParseXMLType int

const (
	VALUES ParseXMLType = iota
	ATTRS
	NAMES
)

func UpdateCacheBySerial(serial string, paramlist []any, l *repository.Cache, xmltype ParseXMLType) {

	device_cache := l.Get(serial)

	if device_cache == nil {
		device_cache = make(map[string]any)
	}

	var new_device_cache map[string]any

	switch xmltype {
	case ATTRS:
		new_device_cache = updateDeviceAttrs(paramlist, device_cache)
	case NAMES:
		new_device_cache = updateDeviceNames(paramlist, device_cache)
	case VALUES:
		new_device_cache = updateDeviceValues(paramlist, device_cache)
	}

	l.Set(serial, new_device_cache)

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
	UpdateCacheBySerial(mp.Serial, paramlist, l, VALUES)

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

	UpdateCacheBySerial(mp.Serial, paramlist, l, NAMES)

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

	UpdateCacheBySerial(mp.Serial, paramlist, l, ATTRS)

	mp.RespChan <- devmodel.SoapResponse{Code: "0", Method: "GetParameterAttributes"}
}
