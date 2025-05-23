package devmodel

import "sync"

type ResponseTask struct {
	RespChan chan any
	Serial   string
	RespList []any
	Body     map[string]any
	mu       sync.RWMutex
}

func InitResponseTask() ResponseTask {
	return ResponseTask{
		RespChan: make(chan any),
		Serial:   "",
		RespList: make([]any, 0, 10),
		Body:     nil,
	}
}

func (r *ResponseTask) InsertRespList(l any) {
	r.mu.Lock()
	r.RespList = append(r.RespList, l)
	r.mu.Unlock()
}
