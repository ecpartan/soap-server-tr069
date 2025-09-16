package usecase_tasks

import (
	"github.com/ecpartan/soap-server-tr069/repository/db/domain/entity"
	"github.com/ecpartan/soap-server-tr069/utils"
)

type Reader interface {
	Get(id utils.ID) (*entity.Task, error)
	Search(query string) ([]*entity.Task, error)
	List() ([]*entity.Task, error)

	GetOp(id utils.ID) (*entity.TaskOp, error)
	SearchOp(query string) ([]*entity.TaskOp, error)
	ListWithOP() ([]*entity.TaskViewDB, error)
}

// Writer user writer
type Writer interface {
	Create(e *entity.Task) (utils.ID, error)
	Update(e *entity.Task) error
	Delete(id utils.ID) error

	CreateOp(e *entity.TaskOp) (utils.ID, error)
	UpdateOp(e *entity.TaskOp) error
	DeleteOp(id utils.ID) error
}

type Repository interface {
	Reader
	Writer
}

type UseCase interface {
	GetTask(id utils.ID) (*entity.Task, error)
	GetAll(limit, offset int) []*entity.Task
	CreateTask(view entity.TaskView) (utils.ID, error)
	UpdateTask(task *entity.Task) error
	DeleteTask(id utils.ID) error

	CreateOperation(op string) (utils.ID, error)
	GetOperation(id utils.ID) (*entity.TaskOp, error)
	GetResultTask(id utils.ID) (string, error)
	GetStatusTask(id utils.ID) (string, error)
	DeleteProfile(id utils.ID) error
}
