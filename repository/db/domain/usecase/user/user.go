package usecase_user

import (
	"github.com/ecpartan/soap-server-tr069/repository/db/domain/entity"
	"github.com/ecpartan/soap-server-tr069/utils"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetUser(id utils.ID) (*entity.User, error) {

	dev, err := s.repo.Get(id)
	if dev == nil {
		return nil, entity.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return dev, nil

}

func (s *Service) GetUserbyLogin(login string) (*entity.User, error) {
	dev, err := s.repo.Search(login)
	if dev == nil {
		return nil, entity.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return dev, nil
}

func (s *Service) GetAll() []*entity.User {
	all, err := s.repo.List()
	if err != nil {
		return nil
	}
	if len(all) == 0 {
		return nil
	}
	return all
}

func (s *Service) CreateUser(view entity.CreateUser) (utils.ID, error) {
	b, err := entity.NewUser(view)
	if err != nil {
		return utils.NewID(), err
	}
	return s.repo.Create(b)
}

func (s *Service) DeleteUser(id utils.ID) error {
	_, err := s.GetUser(id)
	if err != nil {
		return err
	}
	return s.repo.Delete(id)
}

// UpdateBook Update a book
func (s *Service) UpdateUser(d *entity.User) error {
	err := d.Validate()
	if err != nil {
		return err
	}
	return s.repo.Update(d)
}
