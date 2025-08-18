package taskexec

import (
	"sync"

	"github.com/ecpartan/soap-server-tr069/internal/devmodel"
	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/soap-server-tr069/tasks/task"

	"github.com/ecpartan/soap-server-tr069/utils"
)

type ListTasks struct {
	TaskList map[string][]task.Task
	mu       sync.Mutex
}

type TaskExec struct {
	ScripterTasks map[string][]task.Task
	Lst           ListTasks
}

func (e *TaskExec) DeleteTaskByID(serial string, id utils.ID) {
	if maptasks, ok := e.Lst.TaskList[serial]; ok {
		for i, task := range maptasks {
			if task.ID == id {
				e.Lst.mu.Lock()
				defer e.Lst.mu.Unlock()
				e.Lst.TaskList[serial] = append(maptasks[:i], maptasks[i+1:]...)
				break
			}
		}
	}
	logger.LogDebug("DeleteTaskByID", e.Lst.TaskList)
}
func (e *TaskExec) GetListTasksBySerial(serial, host string) []task.Task {
	logger.LogDebug("Lst", e.Lst.TaskList)
	e.Lst.mu.Lock()
	defer e.Lst.mu.Unlock()
	ret_list := e.Lst.TaskList[serial]

	if len(ret_list) == 0 {
		scripterTask := e.findParserTasks(serial)
		if scripterTask != nil {
			return []task.Task{*scripterTask}
		}
	}

	return ret_list
}

func (e *TaskExec) findParserTasks(serial string) *task.Task {
	if tasks, ok := e.ScripterTasks[serial]; ok {
		if len(tasks) < 1 {
			return nil
		}

		ret := tasks[0]
		e.ScripterTasks[serial] = tasks[1:]
		return &ret
	}
	return nil
}

func (e *TaskExec) AddDevicetoTaskList(serial string) {
	e.Lst.mu.Lock()
	if _, ok := e.Lst.TaskList[serial]; !ok {
		e.Lst.TaskList[serial] = []task.Task{}
	}
	e.Lst.mu.Unlock()
}

func (e *TaskExec) CheckNewConReqTasks(mp *devmodel.ResponseTask) {
	logger.LogDebug("CheckNewConReqTasks")
	if script_tasks, ok := e.ScripterTasks[mp.Serial]; ok {
		mp.SetBatchSizeTasks(len(script_tasks))
		e.Lst.TaskList[mp.Serial] = append(e.Lst.TaskList[mp.Serial], script_tasks...)
		e.ScripterTasks[mp.Serial] = []task.Task{}
	}
}
