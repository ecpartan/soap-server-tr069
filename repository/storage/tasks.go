package storage

import (
	"fmt"

	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/soap-server-tr069/repository/db/domain/entity"
	"github.com/ecpartan/soap-server-tr069/utils"
	"github.com/jmoiron/sqlx"
)

type TasksStorage struct {
	db *sqlx.DB
}

func NewTaskstorage(db *sqlx.DB) *TasksStorage {
	return &TasksStorage{db: db}
}

func (s *TasksStorage) Get(id utils.ID) (*entity.Task, error) {
	ret := entity.Task{}

	err := s.db.Select(&ret, "SELECT * FROM task WHERE id = ?", id)

	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (s *TasksStorage) GetOp(id utils.ID) (*entity.TaskOp, error) {
	ret := entity.TaskOp{}

	err := s.db.Select(&ret, "SELECT * FROM task_op WHERE id = ?", id)

	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (s *TasksStorage) GetAll(limit, offset int) []*entity.Task {
	ret := []*entity.Task{}

	err := s.db.Select(&ret, "SELECT * FROM task LIMIT $1 OFFSET $2", limit, offset)

	if err != nil {
		return nil
	}

	return ret
}

func (s *TasksStorage) Search(query string) ([]*entity.Task, error) {

	stmt, err := s.db.Prepare("SELECT * FROM task WHERE")

	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var t []*entity.Task
	rows, err := stmt.Query("%" + query + "%")

	for rows.Next() {
		var cur entity.Task
		if err = rows.Scan(&cur.ID, &cur.Type, &cur.Status, &cur.CreatedAt, &cur.UpdatedAt, &cur.EventCode, &cur.Once, &cur.Result, &cur.TaskOpId, &cur.DeviceId); err != nil {
			return nil, err
		}
		t = append(t, &cur)
	}

	if len(t) == 0 {
		return nil, fmt.Errorf("not found")
	}

	return t, nil
}

func (s *TasksStorage) SearchOp(query string) ([]*entity.TaskOp, error) {

	stmt, err := s.db.Prepare("SELECT * FROM task_op WHERE")

	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var t []*entity.TaskOp
	rows, err := stmt.Query("%" + query + "%")

	for rows.Next() {
		var cur entity.TaskOp
		if err = rows.Scan(&cur.ID, &cur.Body); err != nil {
			return nil, err
		}
		t = append(t, &cur)
	}

	if len(t) == 0 {
		return nil, fmt.Errorf("not found")
	}

	return t, nil
}

func (s *TasksStorage) List() ([]*entity.Task, error) {
	stmt, err := s.db.Query("SELECT * FROM task")

	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var t []*entity.Task

	for stmt.Next() {
		var cur entity.Task
		if err = stmt.Scan(&cur.ID, &cur.Type, &cur.Status, &cur.CreatedAt, &cur.UpdatedAt, &cur.EventCode, &cur.Once, &cur.Result, &cur.TaskOpId, &cur.DeviceId); err != nil {
			return nil, err
		}
		t = append(t, &cur)
	}

	if len(t) == 0 {
		return nil, fmt.Errorf("not found")
	}

	return t, nil
}

func (s *TasksStorage) ListWithOP() ([]*entity.TaskViewDB, error) {
	stmt, err := s.db.Query("SELECT task.id, task.status, task.event_code, task.once, task_op.body FROM task LEFT JOIN task_op ON task.task_op_id = task_op.id;")

	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var t []*entity.TaskViewDB

	for stmt.Next() {
		var cur entity.TaskViewDB
		if err = stmt.Scan(&cur.ID, &cur.Status, &cur.EventCode, &cur.Once, &cur.Body); err != nil {
			return nil, err
		}
		t = append(t, &cur)
	}

	if len(t) == 0 {
		return nil, fmt.Errorf("not found")
	}

	return t, nil
}

func (s *TasksStorage) Create(e *entity.Task) (utils.ID, error) {

	_, err := s.db.Exec("INSERT INTO task (id, type, status, event_code, once, created_at, updated_at, result, task_op_id, device_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
		e.ID, e.Type, e.Status, e.EventCode, e.Once, e.CreatedAt, e.UpdatedAt, e.Result, e.TaskOpId, e.DeviceId)

	logger.LogDebug("Create", err)
	if err != nil {
		return utils.ID{}, err
	}
	return e.ID, nil

}

func (s *TasksStorage) CreateOp(e *entity.TaskOp) (utils.ID, error) {

	k, err := s.db.Exec("INSERT INTO task_op (id, name, body) VALUES ($1, $2, $3)", e.ID, "4444", e.Body)
	logger.LogDebug("CreateOp", k, err)

	if err != nil {
		return utils.ID{}, err
	}
	return e.ID, nil
}

func (s *TasksStorage) Delete(id utils.ID) error {
	return s.db.Get(nil, "DELETE FROM task WHERE id = ?", id)
}

func (s *TasksStorage) DeleteOp(id utils.ID) error {
	return s.db.Get(nil, "DELETE FROM task_op WHERE id = ?", id)
}

func (s *TasksStorage) Update(e *entity.Task) error {

	err := s.db.Get(nil, "UPDATE task SET type = ?, status = ?, event_code = ?, once = ?, result = ?, task_op_id = ?, device_id = ? WHERE id = ?",
		e.Type, e.Status, e.EventCode, e.Once, e.Result, e.TaskOpId, e.DeviceId, e.ID)
	return err
}

func (s *TasksStorage) UpdateOp(e *entity.TaskOp) error {

	err := s.db.Get(nil, "UPDATE task_op SET name = ?, body = ? WHERE id = ?", e.Body, e.ID)

	return err
}
