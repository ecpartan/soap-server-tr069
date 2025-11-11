package task

import (
	"fmt"
	"reflect"

	"github.com/ecpartan/soap-server-tr069/internal/taskmodel"
	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/soap-server-tr069/repository/db/domain/entity"
	"github.com/ecpartan/soap-server-tr069/utils"
)

type TaskRequestType int

const (
	NoTask TaskRequestType = iota
	GetParameterValues
	SetParameterValues
	AddObject
	DeleteObject
	GetParameterNames
	SetParameterAttributes
	GetParameterAttributes
	GetRPCMethods
	Download
	Upload
	FactoryReset
	Reboot
	TransferComplete
)

type Task struct {
	ID        utils.ID
	Action    TaskRequestType
	Params    any
	EventCode int
	Once      bool
}

func createSetParamTask(mapTask []any) []taskmodel.SetParamValTask {

	var settask []taskmodel.SetParamValTask
	settask = make([]taskmodel.SetParamValTask, 0)

	for _, v := range mapTask {

		if iter_map, ok := v.(map[string]any); ok {
			curr_task := taskmodel.SetParamValTask{}
			for k, v := range iter_map {

				switch k {
				case "name":
					curr_task.Name = v.(string)
				case "value":
					curr_task.Value = v.(string)
				case "type":
					curr_task.Type = v.(string)
				}
			}
			settask = append(settask, curr_task)
		}
	}

	fmt.Println("end", settask)

	return settask
}

func createSetParameterAttributesTask(mapTask []any) []taskmodel.SetParamAttrTask {

	var settask []taskmodel.SetParamAttrTask
	settask = make([]taskmodel.SetParamAttrTask, 0)

	for _, v := range mapTask {

		if iter_map, ok := v.(map[string]any); ok {
			curr_task := taskmodel.SetParamAttrTask{}
			for k, v := range iter_map {

				switch k {
				case "name":
					curr_task.Name = v.(string)
				case "notificationChange":
					curr_task.NotificationChange = v.(bool)
				case "notification":
					curr_task.Notification = v.(int)
				case "accessListChange":
					curr_task.AccessListChange = v.(bool)
				case "accessList":
					curr_task.AccessList = v.([]string)
				}
			}
			settask = append(settask, curr_task)
		}
	}

	fmt.Println("end", settask)

	return settask
}

func createDownloadTask(mapTask []any) []taskmodel.DownloadTask {

	var settask []taskmodel.DownloadTask
	settask = make([]taskmodel.DownloadTask, 0)

	for _, v := range mapTask {

		if iter_map, ok := v.(map[string]any); ok {
			curr_task := taskmodel.DownloadTask{}
			for k, v := range iter_map {

				switch k {
				case "filetype":
					curr_task.FileType = v.(string)
				case "url":
					curr_task.URL = v.(string)
				case "username":
					curr_task.Username = v.(string)
				case "password":
					curr_task.Password = v.(string)
				case "delaySeconds":
					curr_task.DelaySeconds = v.(int)
				case "successUrl":
					curr_task.SuccessURL = v.(string)
				case "failureUrl":
					curr_task.FailureURL = v.(string)
				}
			}
			settask = append(settask, curr_task)
		}
	}

	fmt.Println("end", settask)

	return settask
}

func createUploadTask(mapTask []any) []taskmodel.UploadTask {

	var settask []taskmodel.UploadTask
	settask = make([]taskmodel.UploadTask, 0)
	logger.LogDebug("map", mapTask)
	for _, v := range mapTask {

		if iter_map, ok := v.(map[string]any); ok {
			curr_task := taskmodel.UploadTask{}
			logger.LogDebug("map", iter_map)

			for k, v := range iter_map {

				switch k {
				case "filetype":
					curr_task.FileType = v.(string)
				case "url":
					curr_task.URL = v.(string)
				case "username":
					curr_task.Username = v.(string)
				case "password":
					curr_task.Password = v.(string)
				case "delaySeconds":
					curr_task.DelaySeconds = v.(int)
				}
			}
			settask = append(settask, curr_task)
		}
	}

	fmt.Println("end", settask)

	return settask
}

func ParseTask(t map[string]any, tsk *entity.TaskViewDB) *Task {
	logger.LogDebug("parseTask", t)

	for k, v := range t {
		logger.LogDebug("type task", reflect.TypeOf(v), k)
		mapTask, ok := v.(map[string]any)

		if !ok {
			if arrayTask, ok := v.([]any); ok {
				if k == "SetParameterValues" {
					return &Task{
						ID:        utils.NewID(),
						Action:    SetParameterValues,
						Params:    createSetParamTask(arrayTask),
						Once:      true,
						EventCode: 6,
					}
				}
			}
			continue
		}

		rettask := Task{ID: tsk.ID, Once: tsk.Once, EventCode: tsk.EventCode}

		switch k {
		case "AddObject":
			rettask.Action = AddObject
			rettask.Params = taskmodel.AddTask{
				Name: mapTask["Name"].(string),
			}

		case "DeleteObject":
			rettask.Action = DeleteObject
			rettask.Params = taskmodel.DeleteTask{
				Name: mapTask["Name"].(string),
			}

		case "GetParameterValues":
			var lst []string
			if mapN, ok := mapTask["Name"].([]any); ok {
				for _, v := range mapN {
					lst = append(lst, v.(string))
				}
			} else {
				lst = mapTask["Name"].([]string)
			}
			rettask.Action = GetParameterValues
			rettask.Params = taskmodel.GetParamValTask{
				Name: lst,
			}
		case "GetParameterNames":
			logger.LogDebug("GetParameterNames", mapTask["Name"], reflect.TypeOf(mapTask["Name"]))
			logger.LogDebug("GetParameterNames", mapTask["NextLevel"], reflect.TypeOf(mapTask["NextLevel"]))

			if nms, ok := mapTask["Name"].(string); ok {
				if lvls, ok := mapTask["NextLevel"].(float64); ok {
					logger.LogDebug("GetParameterNames", nms, lvls)
					rettask.Action = GetParameterNames
					rettask.Params = taskmodel.GetParamNamesTask{
						ParameterPath: nms,
						NextLevel:     int(lvls),
					}
				}
			}
		case "GetParameterAttributes":
			var lst []string
			if mapN, ok := mapTask["Name"].([]any); ok {
				for _, v := range mapN {
					lst = append(lst, v.(string))
				}
			} else {
				lst = mapTask["Name"].([]string)
			}
			rettask.Action = GetParameterAttributes
			rettask.Params = taskmodel.GetParamAttrTask{
				Name: lst,
			}
		case "SetParameterAttributes":
			if arrayTask, ok := v.([]any); ok {
				rettask.Action = SetParameterAttributes
				rettask.Params = createSetParameterAttributesTask(arrayTask)
			} else {
				return nil
			}
		case "GetRPCMethods":
			rettask.Action = GetRPCMethods
		case "Reboot":
			rettask.Action = Reboot
		case "FactoryReset":
			rettask.Action = FactoryReset
		case "Download":
			rettask.Action = Download
			params := taskmodel.DownloadTask{}

			if nms, ok := mapTask["file"].(string); ok {
				params.FileType = nms
			}
			if nms, ok := mapTask["url"].(string); ok {
				params.URL = nms
			}
			rettask.Params = params
			logger.LogDebug("tsk", rettask)

		case "Upload":
			logger.LogDebug("Upload", v, reflect.TypeOf(v))
			params := taskmodel.UploadTask{}
			if nms, ok := mapTask["file"].(string); ok {
				params.FileType = nms
			}
			if nms, ok := mapTask["url"].(string); ok {
				params.URL = nms
			}
			rettask.Action = Upload
			rettask.Params = params
			logger.LogDebug("tsk", rettask)
		case "TransferComplete":
			rettask.Action = TransferComplete

		}
		return &rettask

	}

	return nil
}
