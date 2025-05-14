package parsemap

import "strings"

func GetXMLValueS(xmlMap any, key string) any {

	strs := strings.Split(key, ".")

	if len(strs) == 0 {
		return nil
	}

	if len(strs) == 1 {
		return GetXMLValue(xmlMap, strs[0])
	}

	for i := range strs {
		xmlMap = GetXMLValue(xmlMap, strs[i])
	}

	return xmlMap
}

func GetXMLValueMap(xmlMap any, key string) map[string]any {
	if mp, ok := xmlMap.(map[string]any); !ok {
		return nil
	} else {
		if value, ok := mp[key]; ok {
			return value.(map[string]any)
		}
	}
	return nil
}

func GetXMLValue(xmlMap any, key string) any {
	if mp, ok := xmlMap.(map[string]any); !ok {
		return nil
	} else {
		if value, ok := mp[key]; ok {
			return value
		}
	}
	return nil
}
