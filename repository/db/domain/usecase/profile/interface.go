package usecase_profile

import (
	"github.com/ecpartan/soap-server-tr069/repository/db/domain/entity"
	"github.com/ecpartan/soap-server-tr069/utils"
)

type Reader interface {
	Get(id utils.ID) (*entity.Profile, error)
	Search(query string) (*entity.Profile, error)
	List() ([]*entity.Profile, error)
}

// Writer user writer
type Writer interface {
	Create(e *entity.Profile) (utils.ID, error)
	Update(e *entity.Profile) error
	Delete(id utils.ID) error
}

type Repository interface {
	Reader
	Writer
}

type UseCase interface {
	GetProfile(id utils.ID) (*entity.Profile, error)
	GetDefaultProfile() (*entity.Profile, error)
	GetAll(limit, offset int) []*entity.Device
	CreateProfile(view entity.ProfileView) (utils.ID, error)
	UpdateProfile(device *entity.Profile) error
	UpdateProfileFw(id utils.ID, profile *entity.Profile) error
	UpdateProfileConf(id utils.ID, profile *entity.Profile) error
	DeleteProfile(id utils.ID) error

	CreateConfig(view entity.ConfigView) (utils.ID, error)
	CreateFirmware(view entity.FirmwareView) (utils.ID, error)
	DeleteConfig(id utils.ID) error
	DeleteFirmware(id utils.ID) error
}
