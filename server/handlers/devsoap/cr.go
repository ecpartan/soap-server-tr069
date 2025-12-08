package devsoap

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ecpartan/soap-server-tr069/httpserver"
	"github.com/ecpartan/soap-server-tr069/internal/apperror"
	"github.com/ecpartan/soap-server-tr069/internal/devmap"
	p "github.com/ecpartan/soap-server-tr069/internal/parsemap"
	logger "github.com/ecpartan/soap-server-tr069/log"
	repository "github.com/ecpartan/soap-server-tr069/repository/cache"
	"github.com/ecpartan/soap-server-tr069/repository/db/domain/entity"
	usecase_device "github.com/ecpartan/soap-server-tr069/repository/db/domain/usecase/device"
	usecase_tasks "github.com/ecpartan/soap-server-tr069/repository/db/domain/usecase/tasks"
	"github.com/ecpartan/soap-server-tr069/server/handlers"
	"github.com/ecpartan/soap-server-tr069/soap"
	"github.com/ecpartan/soap-server-tr069/tasks/scripter"
	"github.com/ecpartan/soap-server-tr069/tasks/tasker"
	"github.com/ecpartan/soap-server-tr069/utils"
	"github.com/julienschmidt/httprouter"
)

type handlerCR struct {
	mapResponse   *devmap.DevMap
	Cache         *repository.Cache
	execTasks     *tasker.Tasker
	taskservice   *usecase_tasks.Service
	deviceService *usecase_device.Service
}

func NewHandlerCR(mapResponse *devmap.DevMap, Cache *repository.Cache, execTasks *tasker.Tasker, taskservice *usecase_tasks.Service, deviceService *usecase_device.Service) handlers.Handler {
	return &handlerCR{
		mapResponse:   mapResponse,
		Cache:         Cache,
		execTasks:     execTasks,
		taskservice:   taskservice,
		deviceService: deviceService,
	}
}

func (h *handlerCR) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodPost, "/addcr", apperror.Middleware(h.PerformConReq))
	router.HandlerFunc(http.MethodPost, "/addtask", apperror.Middleware(h.AddTask))
}

// Connect to the server
// @Summary Perfform a CR
// @Tags SOAP
// @Success 200 {object} tasks.Task
// @Router  /addtask [post]
func (h *handlerCR) PerformConReq(w http.ResponseWriter, r *http.Request) error {

	logger.LogDebug("addtask")
	soapRequestBytes, err := io.ReadAll(r.Body)

	if err != nil {
		return fmt.Errorf("could not read POST: %v", err)
	}

	var getScript map[string]any
	err = json.Unmarshal(soapRequestBytes, &getScript)
	if err != nil {
		return fmt.Errorf("failed unmarshal task CR: %v", err)
	}

	logger.LogDebug("body_task", getScript)

	sn := p.GetSnScript(getScript)
	if sn == "" {
		return fmt.Errorf("failed SN in CR: %v", err)
	}

	err = scripter.AddToScripter(sn, getScript, nil)

	if err != nil {
		return fmt.Errorf("failed add task CR: %v", err)
	}

	tree := h.Cache.Get(sn)

	if tree == nil {
		logger.LogDebug("mp is nil")
		return fmt.Errorf("failed no found SN in DB: %v", err)
	}

	logger.LogDebug("mp", tree)
	url := p.GetXMLValue(tree, soap.CR_URL)
	mp := h.mapResponse.Get(sn)
	user := mp.AuthUsername
	pass := mp.AuthPassword
	if err = httpserver.ExecRequest(url, user, pass); err != nil {
		return err
	}

	return nil

}

type tskType int

const (
	Script tskType = iota
	SetList
	Getlist
	Reboot
	FactoryReset
	FwUpdate
	ConfigUpdate
)

func (t *tskType) String() string {
	switch *t {
	case Script:
		return "Script"
	case SetList:
		return "SetList"
	case Getlist:
		return "Getlist"
	case Reboot:
		return "Reboot"
	case FactoryReset:
		return "FactoryReset"
	case FwUpdate:
		return "FwUpdate"
	case ConfigUpdate:
		return "ConfigUpdate"
	default:
		return "Unknown"
	}
}

func getTaskType(mp map[string]any) (tskType, any) {
	if script_, ok := mp["Script"]; ok {
		return Script, script_
	} else if setList_, ok := mp["SetList"]; ok {
		return SetList, setList_
	} else if getList_, ok := mp["GetList"]; ok {
		return Getlist, getList_
	} else if reboot_, ok := mp["Reboot"]; ok {
		return Reboot, reboot_
	} else if factoryReset_, ok := mp["FactoryReset"]; ok {
		return FactoryReset, factoryReset_
	} else if fwUpdate_, ok := mp["FwUpdate"]; ok {
		return FwUpdate, fwUpdate_
	} else if configUpdate_, ok := mp["ConfigUpdate"]; ok {
		return ConfigUpdate, configUpdate_
	} else {
		return -1, nil
	}
}

// Add task to exec
// @Summary AddTask
// @Tags SOAP
// @Success 200 {object} tasks.Task
// @Router  /addtask [post]
func (h *handlerCR) AddTask(w http.ResponseWriter, r *http.Request) error {

	logger.LogDebug("addtask")
	soapRequestBytes, err := io.ReadAll(r.Body)

	if err != nil {
		return fmt.Errorf("could not read POST: %v", err)
	}

	var mp map[string]any
	err = json.Unmarshal(soapRequestBytes, &mp)
	if err != nil {
		return fmt.Errorf("failed unmarshal task CR: %v", err)
	}

	logger.LogDebug("body_task", mp)
	taskType, taskBody := getTaskType(mp)

	if taskType == -1 {
		return fmt.Errorf("failed task type: %v", err)
	}

	getScript, ok := taskBody.(map[string]any)
	if !ok {
		return fmt.Errorf("failed task body: %v", err)
	}

	sn := p.GetSnScript(mp)
	if sn == "" {
		return fmt.Errorf("failed SN in CR: %v", err)
	}

	logger.LogDebug("body_task", taskBody)

	var oncebool bool
	var eventint int

	if once, ok := mp["Once"]; ok {
		if oncebool, ok = once.(bool); !ok {
			return fmt.Errorf("failed once in CR1: %v", err)
		}
	} else {
		return fmt.Errorf("failed once in CR2: %v", err)
	}

	if event, ok := mp["Event"]; ok {
		if eventf, ok := event.(float64); !ok {
			return fmt.Errorf("failed once in CR: %v", err)
		} else {
			eventint = int(eventf)
		}
	} else {
		return fmt.Errorf("failed once in CR: %v", err)
	}

	view := entity.NewTaskView(taskType.String(), oncebool, eventint)

	str := utils.MapToString(getScript)
	logger.LogDebug("str", str)

	var op_id utils.ID
	if operationId, ok := mp["OperationId"]; !ok {
		if op_id, err = h.taskservice.CreateOperation(str); err != nil {
			return fmt.Errorf("failed create operation: %v", err)
		}
	} else {
		if op_id, err = utils.StringToID(operationId.(string)); err != nil {
			return fmt.Errorf("failed create operation: %v", err)
		}
	}

	dev_id, err := h.deviceService.GetDeviceIDBySn(sn)

	if err != nil {
		return fmt.Errorf("failed sn: %v", err)
	}

	newtsk_id, err := h.taskservice.CreateTask(*view, op_id, dev_id)
	if err != nil {
		return fmt.Errorf("failed create task: %v", err)
	}

	tsk_db := entity.TaskViewDB{ID: newtsk_id, Status: "Pending", Once: oncebool, EventCode: eventint, Body: str}

	switch taskType {
	case Script:
		err = scripter.AddToScripter(sn, getScript, &tsk_db)

		/*
			case SetList:
				err = h.execTasks.AddSetList(taskBody.(map[string]any))
			case Getlist:
				err = h.execTasks.AddGetList(taskBody.(map[string]any))
			case Reboot:
				err = h.execTasks.AddReboot(taskBody.(map[string]any))
			case FactoryReset:
				err = h.execTasks.AddFactoryReset(taskBody.(map[string]any))
			case FwUpdate:
				err = h.execTasks.AddFwUpdate(taskBody.(map[string]any))
			case ConfigUpdate:
				err = h.execTasks.AddConfigUpdate(taskBody.(map[string]any))*/
	default:
		return fmt.Errorf("failed task type: %v", err)
	}

	if err != nil {
		return fmt.Errorf("failed add task CR: %v", err)
	}

	tree := h.Cache.Get(sn)

	if tree == nil {
		logger.LogDebug("mp is nil")
		return fmt.Errorf("failed no found SN in DB: %v", err)
	}

	logger.LogDebug("mp", tree)
	url := p.GetXMLValue(tree, soap.CR_URL)

	m := h.mapResponse.Get(sn)
	user := m.AuthUsername
	pass := m.AuthPassword

	if err = httpserver.ExecRequest(url, user, pass); err != nil {
		return err
	}

	return nil

}
