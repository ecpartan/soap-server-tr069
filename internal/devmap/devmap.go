package devmap

import (
	"sync"

	"github.com/ecpartan/soap-server-tr069/internal/devmodel"
	logger "github.com/ecpartan/soap-server-tr069/log"
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

var (
	d    *DevMap
	once sync.Once
)

func GetDevMap() *DevMap {
	once.Do(func() {
		d = &DevMap{
			devs_session: make(map[string]DevID),
			devs_runtime: make(map[string]DevRun),
			Wg:           &sync.WaitGroup{},
		}
	})
	return d
}

func (d *DevMap) Get(key string) (*DevID, bool) {
	d.RLock()
	defer d.RUnlock()

	ret, ok := d.devs_session[key]
	return &ret, ok
	/*
		if ret, ok := d.devs_session[key]; ok {
			return &ret
		} else {
			tmp := DevID{
				SoapSessionInfo: soap.NewSoapSessionInfo(),
				ResponseTask:    devmodel.NewResponseTask(),
			}
			//d.devs_session[key] = tmp
			return &tmp
		}*/
}

func (d *DevMap) NewSet(key string) *DevID {
	d.Lock()
	defer d.Unlock()

	d.devs_session[key] = DevID{
		SoapSessionInfo: soap.NewSoapSessionInfo(),
		ResponseTask:    devmodel.NewResponseTask(),
	}

	if ret, ok := d.devs_session[key]; ok {
		return &ret
	}

	return nil
}
func (d *DevMap) Set(key string, value DevID) {
	d.Lock()
	defer d.Unlock()
	d.devs_session[key] = value
	logger.LogDebug("Setted", key, d.devs_session)

}

func (d *DevMap) Delete(key string) {
	d.Lock()
	defer d.Unlock()
	if d.devs_session[key].RespChan != nil {
		close(d.devs_session[key].RespChan)
	}
	delete(d.devs_session, key)
	logger.LogDebug("Deleted", key, d.devs_session)
}

func (d *DevMap) GetRuntime(sn string) DevRun {
	d.RLock()
	defer d.RUnlock()
	if ret, ok := d.devs_runtime[sn]; ok {
		return ret
	}
	return DevRun{}
}

func (d *DevMap) SetRuntime(sn string, value DevRun) {
	d.Lock()
	defer d.Unlock()
	d.devs_runtime[sn] = value
}
