package scripter

import (
	"errors"
	"sort"

	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/soap-server-tr069/repository/db/domain/entity"
	"github.com/ecpartan/soap-server-tr069/tasks/task"
	"github.com/ecpartan/soap-server-tr069/tasks/tasker"
	"github.com/ecpartan/soap-server-tr069/utils"
)

func AddToScripter(sn string, scriptList map[string]any, tsk *entity.TaskViewDB) error {

	e := tasker.GetTasker().ExecTasks.Lst.TaskList

	logger.LogDebug("AddToScripter", "scripterTasks", e)

	keys := make([]string, 0, len(scriptList))
	for k := range scriptList {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	if tsk == nil {
		var oncebool bool
		var eventint int

		if once, ok := scriptList["Once"]; ok {
			if oncebool, ok = once.(bool); !ok {
				oncebool = true
			}
		}

		if event, ok := scriptList["Event"]; ok {
			if eventf, ok := event.(float64); !ok {
				eventint = 6
			} else {
				eventint = int(eventf)
			}
		}

		tsk = entity.NewTaskViewDB(utils.NewID(), "Pending", eventint, oncebool, string(utils.MapToString(scriptList)))
	}

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
