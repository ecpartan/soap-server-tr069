package tasks

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ecpartan/soap-server-tr069/httpserver"
	logger "github.com/ecpartan/soap-server-tr069/log"
)

func NextTask(serial, addr string) Task {

	CheckNewConReqTasks(serial, addr)
	tasks := GetListTasksBySerial(serial, addr)

	if len(tasks) == 0 {
		return Task{Action: NoTask}
	}

	eventCodes := s.context.Value("codes").(soap.SetMap)

	for _, task := range tasks {
		if _, ok := eventCodes[task.EventCode]; !ok {
			continue
		}
		switch task.Action {
		case GetParameterValues:
			{
				logger.LogDebug("GetParmeterValues", task)
				DeleteTaskByID(serial, addr, task.ID)

				return task
			}
		case SetParameterValues:
			{
				logger.LogDebug("GetParmeterValues", task)
				DeleteTaskByID(serial, addr, task.ID)

				return task
			}
		case AddObject:
			{
				logger.LogDebug("AddObject", task)
				DeleteTaskByID(serial, addr, task.ID)
				return task
			}
		case DeleteObject:
			{
				logger.LogDebug("DeleteObject", task)
				DeleteTaskByID(serial, addr, task.ID)
				return task
			}
		}
	}

	return Task{Action: NoTask}
}

func executeResponsetask(task_func func(w http.ResponseWriter, req any), task Task, mapResponse *soap.Devmap, w http.ResponseWriter) {

	s.wg.Add(1)
	go func() {
		task_func(w, task.Params)

		if mapResponse.RespChan == nil {
			mapResponse = soap.ResponseTask{
				respChan: make(chan any),
				respList: make([]any, 0),
			}
		}
		ret := <-mapResponse.respChan
		s.log("executeResponsetask", ret)
		respTask := s.mapResponse[host]
		respTask.respList = append(respTask.respList, ret)
		s.mapResponse[host] = respTask
		s.log("executeResponsetask", s.mapResponse)
	}()

	go func() {
		s.wg.Wait()
		s.log("wg.Wait")
	}()
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

func PrepareListTask(task Task, host string) {

	lst := s.mapResponse[host].respList
	s.log("PrepareListTask", lst, len(lst))
	s.log("PrepareListTask", task)
	if len(lst) <= 0 {
		return
	}

	switch task.action {
	case "SetParameterValues":
		{
			s.log("SetParmeterValues")
			task_params := task.params.([]SetParamTask)
			s.log("tasks", task_params)

			for k, v := range task_params {
				str := v.Name
				if ok, start, end := SubstringInstance(str, '#', '.'); ok {
					replacing_trim := str[start:end]
					s.log("replacing_trim", replacing_trim)
					if i, err := strconv.Atoi(replacing_trim[1:]); err == nil {
						if replace_trim, ok := lst[i].(string); ok {
							task_params[k].Name = str[:start] + replace_trim + str[end:]
							s.log("tasks", task_params)
						}
					}
				}
			}
		}
	case "AddObject":
		{
			task_params := task.params.(AddTask)
			str := task_params.Name
			if ok, start, end := SubstringInstance(str, '#', '.'); ok {
				replacing_trim := str[start:end]
				if i, err := strconv.Atoi(replacing_trim[1:]); err == nil {
					if replace_trim, ok := lst[i].(string); ok {
						task_params.Name = str[:start] + replace_trim + str[end:]
					}
				}
			}
		}
	}

}

func ExecuteTask(task Task, host string, w http.ResponseWriter) {

	PrepareListTask(task, host)

	switch task.Action {
	case GetParameterValues:
		{
			executeResponsetask(soap.TransGetParameterValues, task, host, w)
		}
	case SetParameterValues:
		{

			executeResponsetask(httpserver.TransSetParameterValues, task, host, w)
		}

	case AddObject:
		{
			executeResponsetask(httpserver.TransAddObject, task, host, w)
		}

	case DeleteObject:
		{
			executeResponsetask(httpserver.TransDeleteObject, task, host, w)
		}
	}
}

func GetTasks(w http.ResponseWriter, serial, host string) {
	task := NextTask(serial, host)

	if task.Action == NoTask {
		logger.LogDebug("task is nil")
		w.WriteHeader(http.StatusNoContent)
		return
	} else {
		ExecuteTask(task, host, w)
	}
}
