package service

import (
	usecase_device "github.com/ecpartan/soap-server-tr069/repository/db/domain/usecase/device"
	usecase_profile "github.com/ecpartan/soap-server-tr069/repository/db/domain/usecase/profile"
	usecase_tasks "github.com/ecpartan/soap-server-tr069/repository/db/domain/usecase/tasks"
	usecase_user "github.com/ecpartan/soap-server-tr069/repository/db/domain/usecase/user"
	"github.com/ecpartan/soap-server-tr069/repository/storage"
)

type Service struct {
	DeviceService  *usecase_device.Service
	UserService    *usecase_user.Service
	ProfileService *usecase_profile.Service
	TasksService   *usecase_tasks.Service
}

func NewService(s *storage.Storage) *Service {
	return &Service{
		DeviceService:  usecase_device.NewService(s.DevStorage),
		UserService:    usecase_user.NewService(s.UserStorage),
		ProfileService: usecase_profile.NewService(s.ProfileStorage),
		TasksService:   usecase_tasks.NewService(s.TasksStorage),
	}
}
