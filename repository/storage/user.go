package storage

import (
	"github.com/ecpartan/soap-server-tr069/repository/db/domain/entity"
	"github.com/ecpartan/soap-server-tr069/utils"
	"github.com/jmoiron/sqlx"
)

type UserStorage struct {
	db *sqlx.DB
}

func NewUserStorage(db *sqlx.DB) *UserStorage {
	return &UserStorage{db: db}
}

func (s *UserStorage) Get(id utils.ID) (*entity.User, error) {
	ret := entity.User{}

	err := s.db.Select(&ret, "SELECT * FROM users WHERE id = ?", id)

	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (s *UserStorage) GetAll(limit, offset int) []*entity.User {
	ret := []*entity.User{}

	err := s.db.Select(&ret, "SELECT * FROM users LIMIT ? OFFSET ?", limit, offset)

	if err != nil {
		return nil
	}

	return ret
}

func (s *UserStorage) Search(query string) (*entity.User, error) {
	ret := entity.User{}

	err := s.db.Get(&ret, "SELECT * FROM users WHERE sn LIKE ?", "%"+query+"%")

	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (s *UserStorage) List() ([]*entity.User, error) {
	ret := []*entity.User{}
	err := s.db.Select(&ret, "SELECT * FROM users")
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (s *UserStorage) Create(e *entity.User) (utils.ID, error) {

	err := s.db.Get(e, "INSERT INTO users (id, username, email, password, group_id) VALUES (?, ?, ?, ?, ?)", e.ID, e.Username, e.Email, e.Password, e.GroupId)

	if err != nil {
		return utils.ID{}, err
	}
	return e.ID, nil

}

func (s *UserStorage) Delete(id utils.ID) error {

	return s.db.Get(nil, "DELETE FROM users WHERE id = ?", id)
}

func (s *UserStorage) Update(e *entity.User) error {

	err := s.db.Get(nil, "UPDATE users SET username = ?, email = ?, password = ?, group_id = ? WHERE id = ?", e.Username, e.Email, e.Password, e.GroupId, e.ID)
	return err
}
