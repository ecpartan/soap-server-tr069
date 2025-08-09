package service

import (
	usecase_device "github.com/ecpartan/soap-server-tr069/repository/db/domain/usecase/device"
	usecase_user "github.com/ecpartan/soap-server-tr069/repository/db/domain/usecase/user"
)

type Service struct {
	DeviceService *usecase_device.Service
	UserService   *usecase_user.Service
}

func NewService(devRepo usecase_device.Repository, userRepo usecase_user.Repository) *Service {
	return &Service{
		DeviceService: usecase_device.NewService(devRepo),
		UserService:   usecase_user.NewService(userRepo),
	}
}
