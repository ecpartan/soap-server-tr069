package parsemap

import (
	"reflect"
	"strings"

	"github.com/clbanning/mxj/v2"
	logger "github.com/ecpartan/soap-server-tr069/log"
	repository "github.com/ecpartan/soap-server-tr069/repository/cache"
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

func GetXMLValue(xmlMap any, key string) string {
	base := GetXML(xmlMap, key)
	if val, ok := getXML(base, "Value").(string); ok {
		return val
	}
	return ""
}

func ClearCacheNodes(sn string, lst []string) {
	if len(lst) == 0 {
		return
	}
	logger.LogDebug("Clear cache nodes: ", lst)

	tree := repository.GetCache().Get(sn)

	if tree == nil {
		return
	}

	for _, v := range lst {
		DeleteXML(tree, v)
	}

	logger.LogDebug("mp", tree)

	repository.GetCache().Set(sn, tree)
}
func DeleteXML(xmlMap any, key string) {
	strs := strings.Split(key, ".")

	logger.LogDebug("DeleteXML", strs, len(strs))

	if len(strs) == 0 {
		return
	}

	if len(strs) == 1 {
		logger.LogDebug("Delete key: ", strs[0])
		if mp, ok := xmlMap.(map[string]any); ok {
			delete(mp, strs[0])
		}
	}
	for i := 0; i < len(strs)-1; i++ {
		if i == len(strs)-2 {
			logger.LogDebug("mp", reflect.TypeOf(xmlMap))
			logger.LogDebug("mp", xmlMap)
			if mp, ok := xmlMap.(map[string]any); ok {
				logger.LogDebug("Delete key: ", strs[len(strs)-2])
				delete(mp, strs[len(strs)-2])
			}
			break
		}
		currmp := getXML(xmlMap, strs[i])
		xmlMap = currmp

	}
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
