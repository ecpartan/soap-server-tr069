package soap

import "strings"

func GetXMLValueS(xmlMap any, key string) any {

	strs := strings.Split(key, ".")

	if len(strs) == 0 {
		return nil
	}

	if len(strs) == 1 {
		return GetXMLValue(xmlMap, strs[0])
	}

	for i := 0; i < len(strs); i++ {
		xmlMap = GetXMLValue(xmlMap, strs[i])
	}

	return xmlMap
}

func GetXMLValueMap(xmlMap any, key string) map[string]any {
	if parseMap, ok := xmlMap.(map[string]any); !ok {
		return nil
	} else {
		if value, ok := parseMap[key]; ok {
			return value.(map[string]any)
		}
	}
	return nil
}

func GetXMLValue(xmlMap any, key string) any {
	if parseMap, ok := xmlMap.(map[string]any); !ok {
		return nil
	} else {
		if value, ok := parseMap[key]; ok {
			return value
		}
	}
	return nil
}
