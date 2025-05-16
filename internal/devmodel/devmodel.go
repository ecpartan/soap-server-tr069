package devmodel

import "sync"

type ResponseTask struct {
	RespChan chan any
	Serial   string
	RespList []any
	Body     map[string]any
	wg       sync.WaitGroup
}

func InitResponseTask() ResponseTask {
	return ResponseTask{
		RespChan: make(chan any),
		Serial:   "",
		RespList: nil,
		Body:     nil,
	}
}

type DevMap map[string](ResponseTask)

/*
func (d *DevMap) Delete(key string) {
	delete(*d, key)
}

func (d *DevMap) Get(key string) ResponseTask {
	var value ResponseTask
	if value, ok := (*d)[key]; ok {
		return value
	}
	d.Set(key, value)
	return value

}

func (d *DevMap) Set(key string, value ResponseTask) {
	(*d)[key] = value
}

func NewDevMap() *DevMap {
	return &DevMap{}
}
*/
