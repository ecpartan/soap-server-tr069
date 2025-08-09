package usecase_user

import (
	"github.com/ecpartan/soap-server-tr069/repository/db/domain/entity"
	"github.com/ecpartan/soap-server-tr069/utils"
)

// Reader interface
type Reader interface {
	Get(id utils.ID) (*entity.User, error)
	Search(query string) (*entity.User, error)
	List() ([]*entity.User, error)
}

// Writer user writer
type Writer interface {
	Create(e *entity.User) (utils.ID, error)
	Update(e *entity.User) error
	Delete(id utils.ID) error
}

type Repository interface {
	Reader
	Writer
}

// UseCase interface
type UseCase interface {
	GetUser(id utils.ID) (*entity.User, error)
	SearchUsers(query string) ([]*entity.User, error)
	GetAll() []*entity.User
	CreateUser(view entity.CreateUser) (utils.ID, error)
	UpdateUser(e *entity.User) error
	DeleteUser(id utils.ID) error
}
