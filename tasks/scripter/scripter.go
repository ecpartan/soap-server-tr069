package scripter

import (
	"errors"
	"reflect"
	"sort"

	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/soap-server-tr069/repository/db/domain/entity"
	"github.com/ecpartan/soap-server-tr069/tasks/task"
	"github.com/ecpartan/soap-server-tr069/tasks/tasker"
	"github.com/ecpartan/soap-server-tr069/utils"
)

func AddToScripter(sn string, scriptList map[string]any, tsk *entity.TaskViewDB) (int, error) {

	e := tasker.GetTasker().ExecTasks.Lst.TaskList

	logger.LogDebug("AddToScripter", e, scriptList)

	keys := make([]string, 0, len(scriptList))
	for k := range scriptList {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	if tsk == nil {
		oncebool := false
		eventint := 0
		waitevent := -1

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

		if wait, ok := scriptList["WaitEvent"]; ok {
			if waiteventf, ok := wait.(float64); !ok {
				waitevent = -1
			} else {
				waitevent = int(waiteventf)
			}
		}

		tsk = entity.NewTaskViewDB(utils.NewID(), "Pending", eventint, oncebool, string(utils.MapToString(scriptList)), waitevent)
	}

	for _, k := range keys {
		if curr_task, ok := scriptList[k]; ok {
			logger.LogDebug("AddToScripter", reflect.TypeOf(curr_task))
			if addtask, ok := curr_task.(map[string]any); ok {
				find_task := task.ParseTask(sn, addtask, tsk)
				logger.LogDebug("Add task", find_task)
				if find_task == nil {
					return 0, errors.New("failed task")
				}
				e[sn] = append(e[sn], *find_task)
			}
		}
	}
	logger.LogDebug("AddToScripter", "scripterTasks", e)

	return len(e[sn]), nil
}
