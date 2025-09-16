package usecase_profile

import (
	"github.com/ecpartan/soap-server-tr069/repository/db/domain/entity"
	"github.com/ecpartan/soap-server-tr069/utils"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

const df = "default"

func (s *Service) GetProfile(id utils.ID) (*entity.Profile, error) {

	dev, err := s.repo.Get(id)
	if dev == nil {
		return nil, entity.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return dev, nil
}

func (s *Service) GetDefaultProfile() (*entity.Profile, error) {
	profile, err := s.repo.Search(df)
	if profile == nil {
		return nil, entity.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return profile, nil
}

func (s *Service) GetAll(limit, offset int) []*entity.Profile {
	all, err := s.repo.List()
	if err != nil {
		return nil
	}
	if len(all) == 0 {
		return nil
	}
	return all
}

func (s *Service) CreateProfile(view entity.ProfileView) (utils.ID, error) {
	b, err := entity.NewProfile(view)
	if err != nil {
		return utils.NewID(), err
	}
	return s.repo.Create(b)
}

func (s *Service) DeleteProfile(id utils.ID) error {
	_, err := s.GetProfile(id)
	if err != nil {
		return err
	}
	return s.repo.Delete(id)
}

func (s *Service) UpdateProfile(d *entity.Profile) error {
	return s.repo.Update(d)
}

func (s *Service) UpdateProfileFw(id utils.ID, profile *entity.Profile) error {
	profile, err := s.GetProfile(profile.ID)
	if err != nil {
		return err
	}
	profile.FirmwareID = id
	return s.repo.Update(profile)
}

func (s *Service) UpdateProfileConf(id utils.ID, profile *entity.Profile) error {
	profile, err := s.GetProfile(profile.ID)
	if err != nil {
		return err
	}
	profile.ConfigID = id
	return s.repo.Update(profile)
}
