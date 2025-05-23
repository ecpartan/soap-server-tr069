package tasks

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/ecpartan/soap-server-tr069/internal/devmodel"
	"github.com/ecpartan/soap-server-tr069/internal/taskmodel"
	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/soap-server-tr069/soap"
	"github.com/ecpartan/soap-server-tr069/soaprpc"
)

func NextTask(serial, addr string, evcodes map[int]struct{}) Task {

	CheckNewConReqTasks(serial)
	tasks := GetListTasksBySerial(serial, addr)

	if len(tasks) == 0 {
		return Task{Action: NoTask}
	}

	for _, task := range tasks {
		if _, ok := evcodes[task.EventCode]; !ok {
			continue
		}
		switch task.Action {
		case GetParameterValues:
			{
				logger.LogDebug("GetParmeterValues", task)
				DeleteTaskByID(serial, task.ID)

				return task
			}
		case SetParameterValues:
			{
				logger.LogDebug("GetParmeterValues", task)
				DeleteTaskByID(serial, task.ID)

				return task
			}
		case AddObject:
			{
				logger.LogDebug("AddObject", task)
				DeleteTaskByID(serial, task.ID)
				return task
			}
		case DeleteObject:
			{
				logger.LogDebug("DeleteObject", task)
				DeleteTaskByID(serial, task.ID)
				return task
			}
		}
	}

	return Task{Action: NoTask}
}

func SubstringInstance(message string, start, end byte) (bool, int, int) {

	if idx := strings.IndexByte(message, start); idx >= 0 {
		fmt.Println("idx", message[idx:])
		if idx_end := strings.IndexByte(message[idx:], end); idx_end >= 0 {
			return true, idx, idx + idx_end
		} else {
			return true, idx, idx + (idx - len(message) + 1)
		}
	}

	return false, -1, -1
}

func PrepareListTask(task Task, rp *devmodel.ResponseTask) {

	logger.LogDebug("PrepareListTask", rp.RespList)
	logger.LogDebug("PrepareListTask", task)

	if len(rp.RespList) <= 0 {
		return
	}

	switch task.Action {
	case SetParameterValues:
		{
			logger.LogDebug("SetParmeterValues")
			task_params := task.Params.([]taskmodel.SetParamTask)
			logger.LogDebug("tasks", task_params)

			for k, v := range task_params {
				str := v.Name
				if ok, start, end := SubstringInstance(str, '#', '.'); ok {
					replacing_trim := str[start:end]
					logger.LogDebug("replacing_trim", replacing_trim)
					if i, err := strconv.Atoi(replacing_trim[1:]); err == nil {
						if replace_trim, ok := rp.RespList[i].(string); ok {
							task_params[k].Name = str[:start] + replace_trim + str[end:]
							logger.LogDebug("tasks", task_params)
						}
					}
				}
			}
		}
	case AddObject:
		{
			task_params := task.Params.(taskmodel.AddTask)
			str := task_params.Name
			if ok, start, end := SubstringInstance(str, '#', '.'); ok {
				replacing_trim := str[start:end]
				if i, err := strconv.Atoi(replacing_trim[1:]); err == nil {
					if replace_trim, ok := rp.RespList[i].(string); ok {
						task_params.Name = str[:start] + replace_trim + str[end:]
					}
				}
			}
		}
	}

}

func PrepareTasks(host string, mp *devmodel.ResponseTask, sp *soap.SoapResponse) {

	task := NextTask(mp.Serial, host, sp.EventCodes)

	if task.Action == NoTask {
		logger.LogDebug("task is nil")
		return
	} else {
		PrepareListTask(task, mp)
	}
}

func GetTasks(w http.ResponseWriter, host string, mp *devmodel.ResponseTask, sp *soap.SoapResponse, wg *sync.WaitGroup) bool {

	task := NextTask(mp.Serial, host, sp.EventCodes)

	if task.Action == NoTask {
		logger.LogDebug("task is nil")
		w.WriteHeader(http.StatusNoContent)
		return true
	} else {

		PrepareListTask(task, mp)

		ExecuteTask(task, wg, mp, sp, w)
	}
	return false
}

func executeResponsetask(task_func func(w http.ResponseWriter, req any, sp *soap.SoapResponse), task Task, rp *devmodel.ResponseTask, sp *soap.SoapResponse, wg *sync.WaitGroup, w http.ResponseWriter) {
	wg.Add(1)
	go func() {
		task_func(w, task.Params, sp)
		wg.Done()

		ret := <-rp.RespChan
		rp.InsertRespList(ret)
		logger.LogDebug("executeResponsetask", ret)
		logger.LogDebug("executeResponsetask", rp)

	}()

	wg.Wait()
	logger.LogDebug("Wait", task)
}

func ExecuteTask(task Task, wg *sync.WaitGroup, rp *devmodel.ResponseTask, sp *soap.SoapResponse, w http.ResponseWriter) {

	switch task.Action {
	case GetParameterValues:
		{
			executeResponsetask(soaprpc.TransGetParameterValues, task, rp, sp, wg, w)
		}
	case SetParameterValues:
		{

			executeResponsetask(soaprpc.TransSetParameterValues, task, rp, sp, wg, w)
		}

	case AddObject:
		{
			executeResponsetask(soaprpc.TransAddObject, task, rp, sp, wg, w)
		}

	case DeleteObject:
		{
			executeResponsetask(soaprpc.TransDeleteObject, task, rp, sp, wg, w)
		}
	}
}
