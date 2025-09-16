package response

import (
	logger "github.com/ecpartan/soap-server-tr069/log"
)

type RetScriptTask struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

var EndTaskChansMap = make(map[string]chan *RetScriptTask)
var EndTaskResponse = make(map[string]RetScriptTask)

func WriteInChannel(sn string, code string, message string) {
	logger.LogDebug("WriteInChannel", sn, code, message)
	if _, ok := EndTaskChansMap[sn]; !ok {
		EndTaskChansMap[sn] = make(chan *RetScriptTask)
	}
	EndTaskChansMap[sn] <- &RetScriptTask{
		Code:    code,
		Message: message,
	}
	close(EndTaskChansMap[sn])
}

func GetResponse(sn string) *RetScriptTask {
	logger.LogDebug("GetResponse", sn)
	if resp, ok := EndTaskResponse[sn]; !ok {
		return nil
	} else {
		return &resp
	}
}
