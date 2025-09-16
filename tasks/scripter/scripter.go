package scripter

import (
	"errors"
	"sort"

	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/soap-server-tr069/repository/db/domain/entity"
	"github.com/ecpartan/soap-server-tr069/tasks/task"
	"github.com/ecpartan/soap-server-tr069/tasks/tasker"
)

func AddToScripter(sn string, scriptList map[string]any, tsk *entity.TaskViewDB) error {

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
				find_task := task.ParseTask(addtask, tsk)
				logger.LogDebug("Add task", find_task)
				if find_task == nil {
					return errors.New("failed task")
				}
				e[sn] = append(e[sn], *find_task)
			}
		}
	}
	logger.LogDebug("AddToScripter", "scripterTasks", e)

	return nil
}

/*
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
		sn := parsemap.GetSnScript(mp)

		err = AddToScripter(sn, mp, nil)

	}

	logger.LogDebug("Init tasks", exec.ScripterTasks)

	return exec
}
*/
