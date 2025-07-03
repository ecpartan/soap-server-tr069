package devmodel

import (
	"sync"

	"github.com/ecpartan/soap-server-tr069/jrpc2/methods/response"
	logger "github.com/ecpartan/soap-server-tr069/log"
)

type SoapResponse struct {
	Code   string
	Num    string
	Method string
}

type ResponseTask struct {
	RespChan chan SoapResponse
	Serial   string
	RespList []SoapResponse
	BtchSize int
	Body     map[string]any
	mu       sync.RWMutex
}

func NewResponseTask() *ResponseTask {
	return &ResponseTask{
		RespChan: nil,
		Serial:   "",
		RespList: make([]SoapResponse, 0, 10),
		Body:     nil,
	}
}

func (r *ResponseTask) InsertRespList(l SoapResponse) {
	r.mu.Lock()
	r.RespList = append(r.RespList, l)
	logger.LogDebug("InsertRespList", l, r.RespList, r.Serial, r.BtchSize)
	if l.Method == "Fault" {
		response.WriteInChannel(r.Serial, "404", "OK")
		logger.LogDebug("Error")
	} else if len(r.RespList) == r.BtchSize {
		response.WriteInChannel(r.Serial, "200", "OK")
		logger.LogDebug("InsertRespList22", r.RespList, r.Serial, r.BtchSize)
	}

	r.mu.Unlock()
}

func (r *ResponseTask) ResplistIsEmpty() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.RespList) == 0
}

func (r *ResponseTask) SetBatchSizeTasks(sz int) {
	r.mu.Lock()
	if sz > 0 {
		r.BtchSize = sz
	}
	r.mu.Unlock()
}
