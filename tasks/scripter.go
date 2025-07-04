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

func NextTask(mp *devmodel.ResponseTask, addr string, evcodes map[int]struct{}) Task {

	CheckNewConReqTasks(mp)
	tasks := GetListTasksBySerial(mp.Serial, addr)

	if len(tasks) == 0 {
		return Task{Action: NoTask}
	}
	logger.LogDebug("NextTask", tasks)

	for _, task := range tasks {
		if _, ok := evcodes[task.EventCode]; !ok {
			continue
		}
		ret := task
		DeleteTaskByID(mp.Serial, task.ID)

		return ret
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

func PrepareListTask(task Task, rp *devmodel.ResponseTask) Task {

	logger.LogDebug("PrepareListTask", rp.RespList)
	logger.LogDebug("PrepareListTask", task)

	if rp.ResplistIsEmpty() {
		return task
	}

	switch task.Action {
	case SetParameterValues:
		{
			logger.LogDebug("SetParmeterValues")
			task_params := task.Params.([]taskmodel.SetParamValTask)
			logger.LogDebug("tasks", task_params)

			for k, v := range task_params {
				str := v.Name
				if ok, start, end := SubstringInstance(str, '#', '.'); ok {
					replacing_trim := str[start:end]
					logger.LogDebug("replacing_trim", replacing_trim)
					if i, err := strconv.Atoi(replacing_trim[1:]); err == nil {
						replace_trim := rp.RespList[i].Num
						task_params[k].Name = str[:start] + replace_trim + str[end:]
						logger.LogDebug("tasks", task_params)
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
					replace_trim := rp.RespList[i].Num
					task_params.Name = str[:start] + replace_trim + str[end:]
				}
			}
		}
	}
	return task
}

func GetTasks(w http.ResponseWriter, host string, mp *devmodel.ResponseTask, sp *soap.SoapSessionInfo, wg *sync.WaitGroup) bool {
	logger.LogDebug("GetTasks")
	task := NextTask(mp, host, sp.EventCodes)

	if task.Action == NoTask {
		logger.LogDebug("task is nil")

		w.WriteHeader(http.StatusNoContent)

		return true
	} else {
		ExecuteTask(task, wg, mp, sp, w)
	}
	return false
}

var map_tasks = map[TaskRequestType]func(w http.ResponseWriter, req any, sp *soap.SoapSessionInfo){
	GetParameterValues:     soaprpc.TransGetParameterValues,
	SetParameterValues:     soaprpc.TransSetParameterValues,
	AddObject:              soaprpc.TransAddObject,
	DeleteObject:           soaprpc.TransDeleteObject,
	GetParameterNames:      soaprpc.TransGetParameterNames,
	GetParameterAttributes: soaprpc.TransGetParameterAttributes,
}

func executeResponsetask(task_func func(w http.ResponseWriter, req any, sp *soap.SoapSessionInfo), task Task, rp *devmodel.ResponseTask, sp *soap.SoapSessionInfo, wg *sync.WaitGroup, w http.ResponseWriter) {
	logger.LogDebug("executeResponsetask1", rp.RespList)

	wg.Add(1)

	if rp.RespChan == nil {
		rp.RespChan = make(chan devmodel.SoapResponse, 1)
	} else {
		logger.LogDebug("Channel is not empty")
	}

	task = PrepareListTask(task, rp)

	go func(task Task) {
		task_func(w, task.Params, sp)

		wg.Done()

		for val := range rp.RespChan {
			rp.InsertRespList(val)
		}
	}(task)

	logger.LogDebug("executeResponsetask2", rp.RespList)
	logger.LogDebug("Wait", task)

	wg.Wait()
}

func ExecuteTask(task Task, wg *sync.WaitGroup, rp *devmodel.ResponseTask, sp *soap.SoapSessionInfo, w http.ResponseWriter) {
	logger.LogDebug("ExecuteTask", task)

	if action, ok := map_tasks[task.Action]; ok {
		executeResponsetask(action, task, rp, sp, wg, w)
	} else {
		logger.LogDebug("task is nil")
	}
}
