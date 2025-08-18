package tasks

import (
	"fmt"
	"sync"

	"github.com/ecpartan/soap-server-tr069/internal/taskmodel"
	"github.com/ecpartan/soap-server-tr069/tasks/task"
	"github.com/ecpartan/soap-server-tr069/tasks/taskexec"

	"github.com/ecpartan/soap-server-tr069/utils"
)

type Scripter struct {
	tasks           []task.Task
	responsechannel chan task.Task
	mu              sync.Mutex
}

func (s *Scripter) AddTask(task task.Task) {
	s.mu.Lock()
	s.tasks = append(s.tasks, task)
	s.mu.Unlock()
}

func (s *Scripter) RunTasks() {
	for _, task := range s.tasks {
		s.responsechannel <- task
	}
}

func InitTasks() *taskexec.TaskExec {
	stasks := make(map[string][]task.Task)
	lst := make(map[string][]task.Task)
	exec := &taskexec.TaskExec{
		ScripterTasks: stasks,
		Lst: taskexec.ListTasks{
			TaskList: lst,
		},
	}
	paramlistGet := taskmodel.GetParamValTask{}
	paramlistGet.Name = append(paramlistGet.Name, "InternetGatewayDevice.WANDevice.")

	paramsGetName := taskmodel.GetParamNamesTask{
		ParameterPath: "InternetGatewayDevice.WANDevice.",
		NextLevel:     0,
	}
	paramGetAttr := taskmodel.GetParamAttrTask{
		Name: []string{"InternetGatewayDevice.WANDevice."},
	}
	exec.Lst.TaskList["94DE80BF38B2"] = []task.Task{
		{
			ID:        utils.NewID(),
			Action:    task.GetParameterValues,
			Params:    paramlistGet,
			Once:      true,
			EventCode: 1,
		},
		{
			ID:        utils.NewID(),
			Action:    task.GetParameterAttributes,
			Params:    paramGetAttr,
			Once:      true,
			EventCode: 1,
		},
		{
			ID:        utils.NewID(),
			Action:    task.GetParameterNames,
			Params:    paramsGetName,
			Once:      true,
			EventCode: 1,
		},
	}

	fmt.Println(exec.ScripterTasks)

	return exec
}
