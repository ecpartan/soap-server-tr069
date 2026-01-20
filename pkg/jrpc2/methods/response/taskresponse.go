package response

import (
	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/soap-server-tr069/utils"
)

type RetScriptTask struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

var EndTaskChansMap = make(map[string]chan *RetScriptTask)
var EndTaskResponse = make(map[string][]RetScriptTask)

func WriteInChannel(sn string, code string, message map[string]any) {
	logger.LogDebug("WriteInChannel", sn, code, message)
	if _, ok := EndTaskChansMap[sn]; !ok {
		EndTaskChansMap[sn] = make(chan *RetScriptTask, 1)
	}
	EndTaskChansMap[sn] <- &RetScriptTask{
		Code:    code,
		Message: utils.MapToString(message),
	}
}

func CloseChannelBySn(sn string) {
	if _, ok := EndTaskChansMap[sn]; ok {
		close(EndTaskChansMap[sn])
		delete(EndTaskChansMap, sn)
	}
}
