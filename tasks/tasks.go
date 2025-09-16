package tasks

import (
	"encoding/json"
	"sort"
	"sync"

	"github.com/ecpartan/soap-server-tr069/internal/parsemap"
	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/soap-server-tr069/repository/storage"
	"github.com/ecpartan/soap-server-tr069/tasks/task"
	"github.com/ecpartan/soap-server-tr069/tasks/taskexec"
)

type Scripter struct {
	tasks           []task.Task
	responsechannel chan task.Task
	mu              sync.Mutex
}

func InitTasks(s *storage.Storage) *taskexec.TaskExec {
	stasks := make(map[string][]task.Task)
	lst := make(map[string][]task.Task)

	exec := &taskexec.TaskExec{
		ScripterTasks: stasks,
		Lst: taskexec.ListTasks{
			TaskList: lst,
		},
	}
	tskStatorage := s.TasksStorage

	lsts, err := tskStatorage.ListWithOP()
	if err != nil {
		return nil
	}

	for _, tsk := range lsts {

		mp := map[string]any{}
		if err := json.Unmarshal([]byte(tsk.Body), &mp); err != nil {
			logger.LogDebug("Err", err)
		}
		scriptList := parsemap.GetXMLMap(mp, "Script")
		logger.LogDebug("lst", scriptList)
		sn := parsemap.GetSnScript(scriptList)

		keys := make([]string, 0, len(scriptList))
		for k := range scriptList {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			if curr_task, ok := scriptList[k]; ok {
				if addtask, ok := curr_task.(map[string]any); ok {
					find_task := task.ParseTask(addtask, tsk)
					logger.LogDebug("Add task", find_task)
					if find_task == nil {
						continue
					}
					exec.Lst.TaskList[sn] = append(exec.Lst.TaskList[sn], *find_task)
				}
			}
		}
	}

	logger.LogDebug("init list", exec.Lst)

	return exec
}
