package devmap

import (
	"sync"

	"github.com/ecpartan/soap-server-tr069/internal/devmodel"
	"github.com/ecpartan/soap-server-tr069/soap"
)

type DevMap struct {
	sync.RWMutex
	Wg   *sync.WaitGroup
	devs map[string]DevID
}

type DevID struct {
	*devmodel.ResponseTask
	*soap.SoapSessionInfo
}

var d *DevMap

func NewDevMap() *DevMap {
	once := &sync.Once{}
	once.Do(func() {
		d = &DevMap{
			devs: make(map[string]DevID),
			Wg:   &sync.WaitGroup{},
		}
	})
	return d
}

func GetDevMap() *DevMap {
	return d
}
func (d *DevMap) Get(key string) *DevID {
	d.RLock()
	defer d.RUnlock()

	if ret, ok := d.devs[key]; ok {
		return &ret
	} else {
		tmp := DevID{
			SoapSessionInfo: soap.NewSoapSessionInfo(),
			ResponseTask:    devmodel.NewResponseTask(),
		}
		d.devs[key] = tmp
		return &tmp
	}
}
func (d *DevMap) Set(key string, value DevID) {
	d.Lock()
	defer d.Unlock()
	d.devs[key] = value
}

func (d *DevMap) Delete(key string) {
	d.Lock()
	defer d.Unlock()
	if d.devs[key].RespChan != nil {
		close(d.devs[key].RespChan)
	}
	delete(d.devs, key)
}
