package entity

import (
	"time"

	"github.com/ecpartan/soap-server-tr069/utils"
)

type Task struct {
	ID        utils.ID  `db:"id"`
	Type      string    `db:"type"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	Result    string    `db:"result"`
	EventCode int       `db:"event_code"`
	Once      bool      `db:"once"`
	TaskOpId  utils.ID  `db:"task_op_id"`
	DeviceId  utils.ID  `db:"device_id"`
}

type TaskView struct {
	Type      string
	Once      bool
	EventCode int
}

type TaskViewDB struct {
	ID        utils.ID
	Status    string
	EventCode int
	Once      bool
	Body      string
}

func NewTaskViewDB(id utils.ID, status string, eventCode int, once bool, body string) *TaskViewDB {
	return &TaskViewDB{
		ID:        id,
		Status:    status,
		EventCode: eventCode,
		Once:      once,
		Body:      body,
	}
}

func NewTaskView(typ string, once bool, eventCode int) *TaskView {
	return &TaskView{
		Type:      typ,
		Once:      once,
		EventCode: eventCode,
	}
}

func NewTask(t TaskView, opid utils.ID, devid utils.ID) *Task {
	return &Task{
		ID:        utils.NewID(),
		Once:      t.Once,
		EventCode: t.EventCode,
		Status:    "pending",
		TaskOpId:  opid,
		DeviceId:  devid,
	}
}

func (t *Task) Validate() error {
	if t.EventCode < 0 || t.EventCode > 11 {
		return ErrInvalidEntity
	}

	return nil
}

type TaskOp struct {
	ID   utils.ID `db:"id"`
	Body string   `db:"body"`
}

func NewTaskOp(body string) *TaskOp {
	return &TaskOp{
		ID:   utils.NewID(),
		Body: body,
	}
}
