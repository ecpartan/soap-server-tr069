package devmap

import (
	"sync"

	"github.com/ecpartan/soap-server-tr069/internal/devmodel"
	"github.com/ecpartan/soap-server-tr069/soap"
)

type DevMap struct {
	sync.RWMutex
	Wg           *sync.WaitGroup
	devs_session map[string]DevID
	devs_runtime map[string]DevRun
}

type DevID struct {
	*devmodel.ResponseTask
	*soap.SoapSessionInfo
}

type DevRun struct {
	AuthUsername              string
	AuthPassword              string
	ConnectionRequestUsername string
	ConnectionRequestPassword string
}

var d *DevMap

func NewDevMap() *DevMap {
	once := &sync.Once{}
	once.Do(func() {
		d = &DevMap{
			devs_session: make(map[string]DevID),
			devs_runtime: make(map[string]DevRun),
			Wg:           &sync.WaitGroup{},
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

	if ret, ok := d.devs_session[key]; ok {
		return &ret
	} else {
		tmp := DevID{
			SoapSessionInfo: soap.NewSoapSessionInfo(),
			ResponseTask:    devmodel.NewResponseTask(),
		}
		d.devs_session[key] = tmp
		return &tmp
	}
}
func (d *DevMap) Set(key string, value DevID) {
	d.Lock()
	defer d.Unlock()
	d.devs_session[key] = value
}

func (d *DevMap) Delete(key string) {
	d.Lock()
	defer d.Unlock()
	if d.devs_session[key].RespChan != nil {
		close(d.devs_session[key].RespChan)
	}
	delete(d.devs_session, key)
}

func (d *DevMap) GetRuntime(sn string) *DevRun {
	d.RLock()
	defer d.RUnlock()
	if ret, ok := d.devs_runtime[sn]; ok {
		return &ret
	}
	tmp := DevRun{}
	d.devs_runtime[sn] = tmp
	return &tmp
}

func (d *DevMap) SetRuntime(sn string, value DevRun) {
	d.Lock()
	defer d.Unlock()
	d.devs_runtime[sn] = value
}
