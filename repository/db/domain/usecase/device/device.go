package usecase_device

import (
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/ecpartan/soap-server-tr069/repository/db/domain/entity"
	"github.com/ecpartan/soap-server-tr069/utils"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetDevice(id utils.ID) (*entity.Device, error) {

	dev, err := s.repo.Get(id)
	if dev == nil {
		return nil, entity.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return dev, nil
}

func (s *Service) GetOneBySn(sn string) (*entity.Device, error) {
	devs, err := s.repo.Search(sn)
	if devs == nil {
		return nil, entity.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return devs[0], nil
}

func (s *Service) GetDeviceIDBySn(sn string) (utils.ID, error) {
	var id utils.ID
	devs, err := s.repo.Search(sn)
	if devs == nil {
		return id, entity.ErrNotFound
	}

	if err != nil {
		return id, err
	}

	return devs[0].ID, nil
}

func (s *Service) GetIPandPortByID(id utils.ID) (string, int, error) {
	dev, err := s.repo.Get(id)
	if dev == nil {
		return "", 0, entity.ErrNotFound
	}

	if err != nil {
		return "", 0, err
	}

	crurl := dev.CrURL
	parsedURL, err := url.Parse(crurl)
	if err != nil {
		return "", 0, entity.ErrCRULR
	}

	ip, portstr, err := net.SplitHostPort(parsedURL.Host)
	if err != nil {
		return "", 0, entity.ErrCRULR
	}

	port, err := strconv.Atoi(portstr)

	if err != nil {
		return "", 0, entity.ErrCRULR
	}

	return ip, port, nil
}

func (s *Service) GetAll(limit, offset int) []*entity.Device {
	all, err := s.repo.List()
	if err != nil {
		return nil
	}
	if len(all) == 0 {
		return nil
	}
	return all
}

func (s *Service) CreateDevice(view entity.DeviceView) (utils.ID, error) {
	b, err := entity.NewDevice(view)
	if err != nil {
		return utils.NewID(), err
	}
	return s.repo.Create(b)
}

func (s *Service) DeleteDevice(id utils.ID) error {
	_, err := s.GetDevice(id)
	if err != nil {
		return err
	}
	return s.repo.Delete(id)
}

// UpdateBook Update a book
func (s *Service) UpdateDevice(d *entity.Device) error {
	err := d.Validate()
	if err != nil {
		return err
	}
	d.Updated_at = time.Now()
	return s.repo.Update(d)
}

func (s *Service) UpdateDeviceProfileIDbySN(sn string, profileID utils.ID) error {
	devs, err := s.repo.Search(sn)
	if err != nil {
		return err
	}
	d := devs[0]
	d.ProfileId = profileID
	d.Updated_at = time.Now()

	return s.repo.Update(d)
}
