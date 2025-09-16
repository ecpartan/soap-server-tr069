package usecase_tasks

import (
	"time"

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

func (s *Service) GetTask(id utils.ID) (*entity.Task, error) {

	dev, err := s.repo.Get(id)
	if dev == nil {
		return nil, entity.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return dev, nil
}

func (s *Service) GetResultTask(id utils.ID) (string, error) {

	tsk, err := s.GetTask(id)

	if err != nil {
		return "", err
	}

	if tsk == nil {
		return "", entity.ErrNotFound
	}

	return tsk.Result, nil
}

func (s *Service) GetStatusTask(id utils.ID) (string, error) {
	tsk, err := s.repo.Get(id)
	if tsk == nil {
		return "", entity.ErrNotFound
	}

	if err != nil {
		return "", err
	}

	return tsk.Status, nil
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

	dev := devs[0]

	return dev.ID, nil
}

func (s *Service) GetAll(limit, offset int) []*entity.Task {
	all, err := s.repo.List()
	if err != nil {
		return nil
	}
	if len(all) == 0 {
		return nil
	}
	return all
}

func (s *Service) CreateTask(view entity.TaskView, opid utils.ID, devid utils.ID) (utils.ID, error) {
	b := entity.NewTask(view, opid, devid)
	if err := b.Validate(); err != nil {
		return utils.ID{}, err
	}

	b.CreatedAt = time.Now()
	b.UpdatedAt = time.Now()
	b.Result = "CREATED"
	b.Type = "SCRIPT"

	return s.repo.Create(b)
}

func (s *Service) DeleteTask(id utils.ID) error {
	_, err := s.GetTask(id)
	if err != nil {
		return err
	}
	return s.repo.Delete(id)
}

// UpdateBook Update a book
func (s *Service) UpdateTask(t *entity.Task) error {
	err := t.Validate()
	if err != nil {
		return err
	}
	t.UpdatedAt = time.Now()
	return s.repo.Update(t)
}

func (s *Service) UpdateTaskStatus(id utils.ID, status string) error {
	tsk, err := s.GetTask(id)
	if err != nil {
		return err
	}

	tsk.Status = status
	tsk.UpdatedAt = time.Now()
	return s.repo.Update(tsk)
}

func (s *Service) CreateOperation(op string) (utils.ID, error) {
	b := entity.NewTaskOp(op)

	return s.repo.CreateOp(b)
}
