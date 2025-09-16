package tasker

import (
	"github.com/ecpartan/soap-server-tr069/repository/storage"
	"github.com/ecpartan/soap-server-tr069/tasks"
	"github.com/ecpartan/soap-server-tr069/tasks/taskexec"
)

type Tasker struct {
	ExecTasks *taskexec.TaskExec
}

var t *Tasker

func GetTasker() *Tasker {
	return t
}

func InitTasker(s *storage.Storage) *Tasker {
	if t == nil {
		t = &Tasker{ExecTasks: tasks.InitTasks(s)}
	}
	return t
}
