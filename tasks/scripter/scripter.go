package scripter

import (
	"errors"
	"fmt"
	"reflect"
	"sort"

	"github.com/ecpartan/soap-server-tr069/internal/taskmodel"
	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/soap-server-tr069/tasks/task"
	"github.com/ecpartan/soap-server-tr069/tasks/tasker"
	"github.com/ecpartan/soap-server-tr069/utils"
)

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

func parseTask(t map[string]any) *task.Task {
	logger.LogDebug("parseTask", t)

	for k, v := range t {
		logger.LogDebug("type task", reflect.TypeOf(v))
		if mapTask, ok := v.(map[string]any); ok {
			switch k {
			case "AddObject":
				return &task.Task{
					ID:     utils.NewID(),
					Action: task.AddObject,
					Params: taskmodel.AddTask{
						Name: mapTask["Name"].(string),
					},
					Once:      true,
					EventCode: 6,
				}
			case "DeleteObject":
				return &task.Task{
					ID:     utils.NewID(),
					Action: task.DeleteObject,
					Params: taskmodel.DeleteTask{
						Name: mapTask["Name"].(string),
					},
					Once:      true,
					EventCode: 6,
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
				return &task.Task{
					ID:     utils.NewID(),
					Action: task.GetParameterValues,
					Params: taskmodel.GetParamValTask{
						Name: lst,
					},
					Once:      true,
					EventCode: 6,
				}

			case "GetParameterNames":
				return &task.Task{
					ID:     utils.NewID(),
					Action: task.GetParameterNames,
					Params: taskmodel.GetParamNamesTask{
						ParameterPath: mapTask["Name"].(string),
						NextLevel:     mapTask["NextLevel"].(int),
					},
					Once:      true,
					EventCode: 6,
				}
			case "GetParameterAttributes":

				return &task.Task{
					ID:     utils.NewID(),
					Action: task.GetParameterAttributes,
					Params: taskmodel.GetParamAttrTask{
						Name: mapTask["Name"].([]string),
					},
					Once:      true,
					EventCode: 6,
				}
			}
		} else if arrayTask, ok := v.([]any); ok {
			if k == "SetParameterValues" {
				return &task.Task{
					ID:        utils.NewID(),
					Action:    task.SetParameterValues,
					Params:    createSetParamTask(arrayTask),
					Once:      true,
					EventCode: 6,
				}
			}
		}

	}

	return nil
}

func AddToScripter(sn string, scriptList map[string]any) error {

	e := tasker.GetTasker().ExecTasks.Lst.TaskList

	logger.LogDebug("AddToScripter", "scripterTasks", e)

	keys := make([]string, 0, len(scriptList))
	for k := range scriptList {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if curr_task, ok := scriptList[k]; ok {
			if addtask, ok := curr_task.(map[string]any); ok {
				find_task := parseTask(addtask)
				logger.LogDebug("Add task", find_task)
				if find_task == nil {
					return errors.New("failed task")
				}
				e[sn] = append(e[sn], *find_task)
			}
		}
	}
	logger.LogDebug("AddToScripter", "scripterTasks", e)
	//adapter.SetexecTasks(e)

	return nil
}
