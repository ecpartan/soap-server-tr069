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
		}
		return &rettask

	}

	return nil
}
