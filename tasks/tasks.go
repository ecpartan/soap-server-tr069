package tasks

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"sync"

	p "github.com/ecpartan/soap-server-tr069/internal/parsemap"
	"github.com/ecpartan/soap-server-tr069/internal/taskmodel"
	"github.com/ecpartan/soap-server-tr069/utils"
)

type TaskRequestType int

const (
	NoTask TaskRequestType = iota
	GetParameterValues
	SetParameterValues
	AddObject
	DeleteObject
)

type Task struct {
	ID        string
	Action    TaskRequestType
	Params    interface{}
	EventCode int
	Once      bool
}

type TaskResponse struct {
	TaskId       string
	ResponseList any
}

type codeResponse struct {
	Code string
}

type ListTasks struct {
	TaskList map[string][]Task
	mu       sync.Mutex
}

var scripterTasks map[string][]Task

var l ListTasks

type Scripter struct {
	tasks           []Task
	responsechannel chan Task
	mu              sync.Mutex
}

func (s *Scripter) AddTask(task Task) {
	s.mu.Lock()
	s.tasks = append(s.tasks, task)
	s.mu.Unlock()
}

func (s *Scripter) RunTasks() {
	for _, task := range s.tasks {
		s.responsechannel <- task
	}
}

func InitTasks() {

	l.TaskList = make(map[string][]Task)
	scripterTasks = make(map[string][]Task)
	/*
		paramlistSet := []SetParamTask{}
		paramlistSet = append(paramlistSet,
			SetParamTask{Name: "InternetGatewayDevice.WANDevice.1.WANConnectionDevice.1.WANPPPConnection.1.Enable", Value: "1", Type: "xsd:boolean"})
		paramlistSet = append(paramlistSet,
			SetParamTask{Name: "InternetGatewayDevice.WANDevice.1.WANConnectionDevice.1.WANPPPConnection.1.Name", Value: "pppoe_83", Type: "xsd:string"})
	*/

	paramlistGet := taskmodel.GetParamTask{}
	paramlistGet.Name = append(paramlistGet.Name, "InternetGatewayDevice.WANDevice.")

	l.TaskList["94DE80BF38B2"] = []Task{
		{
			ID:        utils.Gen_uuid(),
			Action:    GetParameterValues,
			Params:    paramlistGet,
			Once:      false,
			EventCode: 1,
		},
		{
			ID:        utils.Gen_uuid(),
			Action:    GetParameterValues,
			Params:    paramlistGet,
			Once:      false,
			EventCode: 2,
		},
	}

	/*
		scripterTasks["94DE80BF38B2"] = []Task{
			{
				ID:        utils.Gen_uuid(),
				Action:    GetParameterValues,
				Params:    paramlistGet,
				Once:      false,
				EventCode: 1,
			},
			{
				ID:        utils.Gen_uuid(),
				Action:    GetParameterValues,
				Params:    paramlistGet,
				Once:      false,
				EventCode: 2,
			},
		}*/
	fmt.Println(scripterTasks)

}

/*
	func AddTask(serial string, task Task) {
		if _, ok := l.TaskList[serial]; ok {
			l.mu.Lock()
			defer l.mu.Unlock()
			l.TaskList[serial] = append(l.TaskList[serial], task)
		} else {
			l.TaskList[serial] = []Task{task}
		}
	}
*/
func DeleteTaskByID(serial, id string) {
	if maptasks, ok := l.TaskList[serial]; ok {
		for i, task := range maptasks {
			if task.ID == id {
				l.mu.Lock()
				defer l.mu.Unlock()
				l.TaskList[serial] = append(maptasks[:i], maptasks[i+1:]...)
				break
			}
		}
	}
	fmt.Println("DeleteTaskByID", l.TaskList)
}
func GetListTasksBySerial(serial, host string) []Task {
	fmt.Println(l.TaskList)
	l.mu.Lock()
	defer l.mu.Unlock()
	ret_list := l.TaskList[serial]
	if len(ret_list) == 0 {
		scripterTask := findParserTasks(serial)
		if scripterTask != nil {
			return []Task{*scripterTask}
		}
	}

	return ret_list
}

func findParserTasks(serial string) *Task {
	if tasks, ok := scripterTasks[serial]; ok {
		if len(tasks) < 1 {
			return nil
		}

		ret := tasks[0]
		scripterTasks[serial] = tasks[1:]
		return &ret
	}
	return nil
}

func AddDevicetoTaskList(serial string) {
	l.mu.Lock()
	if _, ok := l.TaskList[serial]; !ok {
		l.TaskList[serial] = []Task{}
	}
	l.mu.Unlock()
}

func createSetParamTask(mapTask []any) []taskmodel.SetParamTask {

	var settask []taskmodel.SetParamTask
	settask = make([]taskmodel.SetParamTask, 0)

	for _, v := range mapTask {

		if iter_map, ok := v.(map[string]any); ok {
			curr_task := taskmodel.SetParamTask{}
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

func parseTask(task map[string]any) (*Task, error) {
	fmt.Println(task)
	for k, v := range task {
		fmt.Println(reflect.TypeOf(v))
		if mapTask, ok := v.(map[string]any); ok {
			switch k {
			case "AddObject":
				return &Task{
					ID:     utils.Gen_uuid(),
					Action: AddObject,
					Params: taskmodel.AddTask{
						Name: mapTask["Name"].(string),
					},
					Once:      true,
					EventCode: 6,
				}, nil
			case "DeleteObject":
				return &Task{
					ID:     utils.Gen_uuid(),
					Action: DeleteObject,
					Params: taskmodel.DeleteTask{
						Name: mapTask["Name"].(string),
					},
					Once:      true,
					EventCode: 6,
				}, nil
			case "GetParameterValues":
				return &Task{
					ID:     utils.Gen_uuid(),
					Action: GetParameterValues,
					Params: taskmodel.GetParamTask{
						Name: mapTask["Name"].([]string),
					},
					Once:      true,
					EventCode: 6,
				}, nil
			}
		} else if arrayTask, ok := v.([]any); ok {
			if k == "SetParameterValues" {
				return &Task{
					ID:        utils.Gen_uuid(),
					Action:    SetParameterValues,
					Params:    createSetParamTask(arrayTask),
					Once:      true,
					EventCode: 6,
				}, nil
			}
		}

	}

	return nil, errors.New("Task is invalid")
}
func ParseScriptToTask(getScript map[string]any) (string, error) {
	script := p.GetXMLValue(getScript, "Script")

	if script == nil {
		return "", errors.New("script is empty")
	}

	serial, ok := p.GetXMLValue(script, "Serial").(string)
	if !ok || serial == "" {
		return "", errors.New("serial is empty")
	}

	if scriptList, ok := script.(map[string]any); ok {
		keys := make([]string, 0, len(scriptList))
		for k := range scriptList {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			if curr_task, ok := scriptList[k]; ok {
				if addtask, ok := curr_task.(map[string]any); ok {
					find_task, err := parseTask(addtask)

					if err != nil {
						return "", err
					}
					scripterTasks[serial] = append(scripterTasks[serial], *find_task)
				}
			}
		}
	}

	return serial, nil
}
func CheckNewConReqTasks(serial string) {
	if script_tasks, ok := scripterTasks[serial]; ok {
		l.TaskList[serial] = append(l.TaskList[serial], script_tasks...)
		scripterTasks[serial] = []Task{}
	}
}
