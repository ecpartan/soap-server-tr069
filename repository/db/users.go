package db

import (
	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/soap-server-tr069/repository/db/dao"
)

func (s *Service) GetUsers() ([]dao.User, error) {
	ret := make([]dao.User, 0)

	err := s.db.Select(&ret, "SELECT id, username,password FROM user")
	logger.LogDebug("GetUsers", ret, err)
	if err != nil {
		return nil, err
	}

	logger.LogDebug("GetUsers", ret)

	return ret, err
}

func (s *Service) GetUser(username string) (dao.User, error) {
	ret := dao.User{}
	err := s.db.Get(&ret, "SELECT id, user,password FROM user WHERE username = ?", username)
	if err != nil {
		return dao.User{}, err
	}
	return ret, err
}
