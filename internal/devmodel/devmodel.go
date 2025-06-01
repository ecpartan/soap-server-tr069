package devmodel

import "sync"

type ResponseTask struct {
	RespChan chan any
	Serial   string
	RespList []any
	Body     map[string]any
	mu       sync.RWMutex
}

func NewResponseTask() *ResponseTask {
	return &ResponseTask{
		RespChan: nil,
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

func (r *ResponseTask) ResplistIsEmpty() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.RespList) == 0
}
