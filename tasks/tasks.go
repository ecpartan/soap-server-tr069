package tasks

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"sync"

	p "github.com/ecpartan/soap-server-tr069/internal/parsemap"
	"github.com/ecpartan/soap-server-tr069/utils"
)

type TaskRequestType int

const (
	GetParameterValuesR TaskRequestType = iota
	SetParameterValuesR
	AddObjectR
	DeleteObjectR
	NoTaskRequestR
)

type Task struct {
	id        string
	action    string
	params    interface{}
	eventCode int
	once      bool
}

type TaskResponse struct {
	TaskId       string
	ResponseList any
}

type codeResponse struct {
	Code string
}

/*
	type AddObject struct {
		codeResponse
		AddInstanse string
	}
*/
type deviceid struct {
	serial string
	host   string
}

type ListTasks struct {
	TaskList map[deviceid][]Task
	mu       sync.Mutex
}

var scripterTasks map[string][]Task

var l ListTasks

type Scripter struct {
	tasks           []Task
	responsechannel chan Task
}

func (s *Scripter) AddTask(task Task) {
	s.tasks = append(s.tasks, task)
}

func (s *Scripter) AddTaskList(tasklist []Task) {
	s.tasks = append(s.tasks, tasklist...)
}

func (s *Scripter) RunTasks() {
	for _, task := range s.tasks {
		s.responsechannel <- task
	}
}

type SetParamTask struct {
	Name  string
	Value string
	Type  string
}

type GetParamTask struct {
	Name []string
}

type AddTask struct {
	Name string
}

type DeleteTask struct {
	Name string
}

type SoapTask interface {
	AddTask | DeleteTask | GetParamTask | SetParamTask | []SetParamTask
}

func InitTasks() {

	l.TaskList = make(map[deviceid][]Task)
	scripterTasks = make(map[string][]Task)
	/*
		paramlistSet := []SetParamTask{}
		paramlistSet = append(paramlistSet,
			SetParamTask{Name: "InternetGatewayDevice.WANDevice.1.WANConnectionDevice.1.WANPPPConnection.1.Enable", Value: "1", Type: "xsd:boolean"})
		paramlistSet = append(paramlistSet,
			SetParamTask{Name: "InternetGatewayDevice.WANDevice.1.WANConnectionDevice.1.WANPPPConnection.1.Name", Value: "pppoe_83", Type: "xsd:string"})
	*/
	/*
		paramlistGet := GetParamTask{}
		paramlistGet.Name = append(paramlistGet.Name, "InternetGatewayDevice.WANDevice.")


		l.TaskList[deviceid{serial: "94DE80BF38B2", host: "127.0.0.1:8089"}] = []Task{
			{
				id:        gen_uuid(),
				action:    "GetParmeterValues",
				params:    paramlistGet,
				once:      false,
				eventCode: 1,
			},
			{
				id:        gen_uuid(),
				action:    "GetParmeterValues",
				params:    paramlistGet,
				once:      false,
				eventCode: 2,
			},

		}
		fmt.Println(l.TaskList)
	*/
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
func DeleteTaskByID(serial, host, id string) {
	deviceID := deviceid{serial: serial, host: host}
	if maptasks, ok := l.TaskList[deviceID]; ok {
		for i, task := range maptasks {
			if task.id == id {
				l.mu.Lock()
				defer l.mu.Unlock()
				l.TaskList[deviceID] = append(maptasks[:i], maptasks[i+1:]...)
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
	ret_list := l.TaskList[deviceid{serial: serial, host: host}]
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

func AddDevicetoTaskList(serial, addr string) {
	id := deviceid{serial: serial, host: addr}
	l.mu.Lock()
	if _, ok := l.TaskList[id]; !ok {
		l.TaskList[id] = []Task{}
	}
	l.mu.Unlock()
}

func createSetParamTask(mapTask []any) []SetParamTask {

	var settask []SetParamTask
	settask = make([]SetParamTask, 0)

	for _, v := range mapTask {

		if iter_map, ok := v.(map[string]any); ok {
			curr_task := SetParamTask{}
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
					id:     utils.Gen_uuid(),
					action: "AddObject",
					params: AddTask{
						Name: mapTask["Name"].(string),
					},
					once:      true,
					eventCode: 6,
				}, nil
			case "DeleteObject":
				return &Task{
					id:     utils.Gen_uuid(),
					action: "DeleteObject",
					params: DeleteTask{
						Name: mapTask["Name"].(string),
					},
					once:      true,
					eventCode: 6,
				}, nil
			case "GetParameterValues":
				return &Task{
					id:     utils.Gen_uuid(),
					action: "GetParameterValues",
					params: GetParamTask{
						Name: mapTask["Name"].([]string),
					},
					once:      true,
					eventCode: 6,
				}, nil
			}
		} else if arrayTask, ok := v.([]any); ok {
			if k == "SetParameterValues" {
				return &Task{
					id:        utils.Gen_uuid(),
					action:    "SetParameterValues",
					params:    createSetParamTask(arrayTask),
					once:      true,
					eventCode: 6,
				}, nil
			}
		}

	}

	return nil, errors.New("Task is invalid")
}
func ParseScriptToTask(getScript map[string]any) error {
	script := p.GetXMLValue(getScript, "Script")

	if script == nil {
		return errors.New("Script is empty")
	}

	serial := p.GetXMLValue(script, "Serial").(string)
	if serial == "" {
		return errors.New("Serial is empty")
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
						return err
					}
					scripterTasks[serial] = append(scripterTasks[serial], *find_task)
				}
			}

		}
	}

	return nil
}

func CheckNewConReqTasks(serial, host string) {
	id := deviceid{serial: serial, host: host}
	if script_tasks, ok := scripterTasks[serial]; ok {
		l.TaskList[id] = append(l.TaskList[id], script_tasks...)
		scripterTasks[serial] = []Task{}
	}
}
