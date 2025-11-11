package tasks

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/ecpartan/soap-server-tr069/httpserver"
	"github.com/ecpartan/soap-server-tr069/internal/devmodel"
	p "github.com/ecpartan/soap-server-tr069/internal/parsemap"
	"github.com/ecpartan/soap-server-tr069/internal/taskmodel"
	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/soap-server-tr069/soap"
	"github.com/ecpartan/soap-server-tr069/tasks/task"
	"github.com/ecpartan/soap-server-tr069/tasks/taskexec"
)

func NextTask(mp *devmodel.ResponseTask, addr string, evcodes map[int]struct{}, e *taskexec.TaskExec) task.Task {

	e.CheckNewConReqTasks(mp)
	tasks := e.GetListTasksBySerial(mp.Serial, addr)

	if len(tasks) == 0 {
		return task.Task{Action: task.NoTask}
	}
	logger.LogDebug("NextTask", tasks)

	for _, task := range tasks {
		if _, ok := evcodes[task.EventCode]; !ok {
			continue
		}
		ret := task
		e.DeleteTaskByID(mp.Serial, task.ID)

		return ret
	}

	return task.Task{Action: task.NoTask}
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

func PrepareListTask(t task.Task, rp *devmodel.ResponseTask) task.Task {

	logger.LogDebug("PrepareListTask", rp.RespList)
	logger.LogDebug("PrepareListTask", t)

	if len(rp.RespList) == 0 {
		return t
	}

	switch t.Action {
	case task.SetParameterValues:
		{
			logger.LogDebug("SetParmeterValues")
			task_params := t.Params.([]taskmodel.SetParamValTask)
			logger.LogDebug("tasks", task_params)

			for k, v := range task_params {
				str := v.Name
				if ok, start, end := SubstringInstance(str, '#', '.'); ok {
					replacing_trim := str[start:end]
					logger.LogDebug("replacing_trim", replacing_trim)
					if i, err := strconv.Atoi(replacing_trim[1:]); err == nil && i < len(rp.RespList) {
						replace_trim := rp.RespList[i].Num
						task_params[k].Name = str[:start] + replace_trim + str[end:]
						logger.LogDebug("tasks", task_params)
					}
				}
			}
		}
	case task.AddObject:
		{
			task_params := t.Params.(taskmodel.AddTask)
			str := task_params.Name
			if ok, start, end := SubstringInstance(str, '#', '.'); ok {
				replacing_trim := str[start:end]
				if i, err := strconv.Atoi(replacing_trim[1:]); err == nil && i < len(rp.RespList) {
					replace_trim := rp.RespList[i].Num
					task_params.Name = str[:start] + replace_trim + str[end:]
				}
			}
		}
	}
	return t
}

func GetTasks(w http.ResponseWriter, host string, mp *devmodel.ResponseTask, sp *soap.SoapSessionInfo, wg *sync.WaitGroup, e *taskexec.TaskExec) bool {
	logger.LogDebug("GetTasks")
	t := NextTask(mp, host, sp.EventCodes, e)

	if t.Action == task.NoTask {
		logger.LogDebug("task is nil")

		w.WriteHeader(http.StatusNoContent)

		return true
	} else {
		if t.Action == task.GetParameterValues {
			logger.LogDebug("GetParameterValues", t.Params.(taskmodel.GetParamValTask))
			p.ClearCacheNodes(mp.Serial, t.Params.(taskmodel.GetParamValTask).Name)
		}
		ExecuteTask(t, wg, mp, sp, w)
	}
	return false
}

var map_tasks = map[task.TaskRequestType]func(w http.ResponseWriter, req any, sp *soap.SoapSessionInfo){
	task.GetParameterValues:     httpserver.TransGetParameterValues,
	task.SetParameterValues:     httpserver.TransSetParameterValues,
	task.AddObject:              httpserver.TransAddObject,
	task.DeleteObject:           httpserver.TransDeleteObject,
	task.GetParameterNames:      httpserver.TransGetParameterNames,
	task.GetParameterAttributes: httpserver.TransGetParameterAttributes,
	task.SetParameterAttributes: httpserver.TransSetParameterAttributes,
	task.Download:               httpserver.TransDownload,
	task.Upload:                 httpserver.TransUpload,
	task.Reboot:                 httpserver.TransReboot,
	task.FactoryReset:           httpserver.TransFactoryReset,
	task.GetRPCMethods:          httpserver.TransGetRPCMethods,
	task.TransferComplete:       httpserver.TransTransferCompleteResponse,
}

func executeResponsetask(task_func func(w http.ResponseWriter, req any, sp *soap.SoapSessionInfo), t task.Task, rp *devmodel.ResponseTask, sp *soap.SoapSessionInfo, wg *sync.WaitGroup, w http.ResponseWriter) {
	logger.LogDebug("executeResponsetask1", rp.RespList)

	wg.Add(1)

	if rp.RespChan == nil {
		rp.RespChan = make(chan devmodel.SoapResponse)
	} else {
		close(rp.RespChan)
		rp.RespChan = make(chan devmodel.SoapResponse)
		logger.LogDebug("Channel is not empty")
	}

	t = PrepareListTask(t, rp)
	logger.LogDebug("executeResponsetask2", t)
	go func(t task.Task) {
		logger.LogDebug("executeResponsetask2", t)
		task_func(w, t.Params, sp)

		wg.Done()

		for val := range rp.RespChan {
			rp.InsertRespList(val)
		}
	}(t)

	logger.LogDebug("executeResponsetask2", rp.RespList)
	logger.LogDebug("Wait", t)

	wg.Wait()
}

func ExecuteTask(task task.Task, wg *sync.WaitGroup, rp *devmodel.ResponseTask, sp *soap.SoapSessionInfo, w http.ResponseWriter) {
	logger.LogDebug("ExecuteTask", task)

	if action, ok := map_tasks[task.Action]; ok {
		executeResponsetask(action, task, rp, sp, wg, w)
	} else {
		logger.LogDebug("task is nil")
	}
}
