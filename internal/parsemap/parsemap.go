package parsemap

import (
	"strings"

	"github.com/clbanning/mxj/v2"
)

func ConvertXMLtoMap(in []byte) (map[string]any, error) {
	var mv map[string]any
	mv, err := mxj.NewMapXmlSeq(in)
	if err != nil {
		return nil, err
	}
	return mv, nil
}

func GetXML(xmlMap any, key string) any {
	strs := strings.Split(key, ".")

	if len(strs) == 0 {
		return nil
	}

	if len(strs) == 1 {
		return getXML(xmlMap, strs[0])
	}

	for i := range strs {
		xmlMap = getXML(xmlMap, strs[i])
	}

	return xmlMap
}

func GetXMLString(xmlMap any, key string) string {
	if mp := GetXML(xmlMap, key); mp != nil {
		if str, ok := GetXML(mp, "#text").(string); ok {
			return str
		}
	}

	return ""
}

func getXML(xmlMap any, key string) any {
	if mp, ok := xmlMap.(map[string]any); ok {
		if value, ok := mp[key]; ok {
			return value
		}
	}
	if mp, ok := xmlMap.(map[string]string); ok {
		if value, ok := mp[key]; ok {
			return value
		}
	}
	return nil
}

func GetSnScript(getScript map[string]any) string {

	if sn, ok := GetXML(getScript, "Serial").(string); ok {
		return sn
	}
	return ""
}
