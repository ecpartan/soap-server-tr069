package tasks

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ecpartan/soap-server-tr069/soap"
)

func NextTask(serial, addr string) (TaskRequestType, Task) {

	CheckNewConReqTasks(serial, addr)
	tasks := GetListTasksBySerial(serial, addr)

	if len(tasks) == 0 {
		return NoTaskRequestR, Task{}
	}

	eventCodes := s.context.Value("codes").(soap.SetMap)

	for _, task := range tasks {
		if _, ok := eventCodes[task.eventCode]; !ok {
			continue
		}
		s.log("Action", task.action)
		switch task.action {
		case "GetParmeterValues":
			{
				s.log("GetParmeterValues")
				DeleteTaskByID(serial, addr, task.id)

				return GetParameterValuesR, task
			}
		case "SetParameterValues":
			{
				s.log("SetParmeterValues")
				s.log(task)
				DeleteTaskByID(serial, addr, task.id)
				s.log(task)

				return SetParameterValuesR, task
			}
		case "AddObject":
			{
				s.log("AddObject")
				DeleteTaskByID(serial, addr, task.id)
				return AddObjectR, task
			}
		case "DeleteObject":
			{
				s.log("DeleteObject")
				DeleteTaskByID(serial, addr, task.id)
				return DeleteObjectR, task
			}
		}
	}

	return NoTaskRequestR, Task{}
}

func ExecuteResponsetask(task_func func(w http.ResponseWriter, req any), task Task, host string, w http.ResponseWriter) {

	s.wg.Add(1)
	go func() {
		s.log(task)
		task_func(w, task.params)

		if s.mapResponse[host].respChan == nil {
			s.mapResponse[host] = responseTask{
				respChan: make(chan any),
				respList: make([]any, 0),
			}
		}
		ret := <-s.mapResponse[host].respChan
		s.log("executeResponsetask", ret)
		respTask := s.mapResponse[host]
		respTask.respList = append(respTask.respList, ret)
		s.mapResponse[host] = respTask
		s.log("executeResponsetask", s.mapResponse)
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

func ExecuteTask(Action TaskRequestType, task Task, host string, w http.ResponseWriter) {

	if task.action == "" {
		return
	}

	switch Action {
	case GetParameterValuesR:
		{

			s.executeResponsetask(s.TransGetParameterValues, task, host, w)
		}
	case SetParameterValuesR:
		{

			PrepareListTask(task, host)
			s.executeResponsetask(s.TransSetParameterValues, task, host, w)
		}

	case AddObjectR:
		{
			s.log("AddObjectR")
			s.executeResponsetask(s.TransAddObject, task, host, w)
		}

	case DeleteObjectR:
		{
			s.log("DeleteObjectR")
			s.executeResponsetask(s.TransDeleteObject, task, host, w)
		}
	}
	s.wg.Wait()
	s.log("ExecuteTask end")
}

func GetTasks(w http.ResponseWriter, host string) {
	var serial string
	if deviceID, ok := s.context.Value("DeviceID").(DeviceId); ok {
		s.log("deviceID", deviceID)
		serial = deviceID.SerialNumber
	}
	if serial == "" {
		s.log("serial is empty")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	taskAction, task := NextTask(serial, host)

	if taskAction == NoTaskRequestR {
		s.log("task is nil")
		w.WriteHeader(http.StatusNoContent)
		return
	} else {
		s.ExecuteTask(taskAction, task, host, w)
	}
}
