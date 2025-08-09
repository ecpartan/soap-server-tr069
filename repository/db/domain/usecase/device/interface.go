package usecase_device

import (
	"github.com/ecpartan/soap-server-tr069/repository/db/domain/entity"
	"github.com/ecpartan/soap-server-tr069/utils"
)

// Reader interface
type Reader interface {
	Get(id utils.ID) (*entity.Device, error)
	Search(query string) (*entity.Device, error)
	List() ([]*entity.Device, error)
}

// Writer user writer
type Writer interface {
	Create(e *entity.Device) (utils.ID, error)
	Update(e *entity.Device) error
	Delete(id utils.ID) error
}

type Repository interface {
	Reader
	Writer
}

type UseCase interface {
	GetDevice(id utils.ID) (*entity.Device, error)
	GetOneBySn(sn string) (*entity.Device, error)
	GetAll(limit, offset int) []*entity.Device
	CreateDevice(view entity.DeviceView) (utils.ID, error)
	UpdateDevice(device *entity.Device) error
	DeleteDevice(id utils.ID) error
}
