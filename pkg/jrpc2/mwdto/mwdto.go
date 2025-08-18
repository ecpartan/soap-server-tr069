package mwdto

import "github.com/ecpartan/soap-server-tr069/tasks/taskexec"

type Mwdto struct {
	Reqw      map[string]any
	ExecTasks *taskexec.TaskExec
}
