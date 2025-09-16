package storage

import (
	"fmt"

	"github.com/ecpartan/soap-server-tr069/repository/db/domain/entity"
	"github.com/ecpartan/soap-server-tr069/utils"
	"github.com/jmoiron/sqlx"
)

type ProfileStorage struct {
	db *sqlx.DB
}

func NewProfileStorage(db *sqlx.DB) *ProfileStorage {
	return &ProfileStorage{db: db}
}

func (s *ProfileStorage) Get(id utils.ID) (*entity.Profile, error) {
	ret := entity.Profile{}

	err := s.db.Select(&ret, "SELECT * FROM profile WHERE id = ?", id)

	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (s *ProfileStorage) GetAll(limit, offset int) []*entity.Profile {
	ret := []*entity.Profile{}

	err := s.db.Select(&ret, "SELECT * FROM profile LIMIT $1 OFFSET $2", limit, offset)

	if err != nil {
		return nil
	}

	return ret
}

func (s *ProfileStorage) Search(query string) (*entity.Profile, error) {
	ret := entity.Profile{}

	stmt, err := s.db.Prepare("SELECT * FROM profile WHERE")

	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var ids []utils.ID

	rows, err := stmt.Query("%" + query + "%")

	for rows.Next() {
		var i utils.ID
		err = rows.Scan(&i)
		if err != nil {
			return nil, err
		}
		ids = append(ids, i)
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("not found")
	}

	return &ret, nil
}

func (s *ProfileStorage) List() ([]*entity.Profile, error) {
	ret := []*entity.Profile{}
	err := s.db.Select(&ret, "SELECT * FROM profile")
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (s *ProfileStorage) Create(e *entity.Profile) (utils.ID, error) {

	err := s.db.Get(e, "INSERT INTO profile (id, name, description, firmware_id, config_id) VALUES (?, ?, ?, ?, ?) RETURNING id",
		e.ID, e.Name, e.Description, e.FirmwareID, e.ConfigID)

	if err != nil {
		return utils.ID{}, err
	}
	return e.ID, nil

}

func (s *ProfileStorage) CreateFW(e *entity.Firmware) (utils.ID, error) {

	err := s.db.Get(e, "INSERT INTO firmware (id, name, path, size, version) VALUES (?, ?, ?, ?, ?) RETURNING id",
		e.ID, e.Name, e.Path, e.Size, e.Version)

	if err != nil {
		return utils.ID{}, err
	}
	return e.ID, nil
}

func (s *ProfileStorage) CreateConf(e *entity.Config) (utils.ID, error) {

	err := s.db.Get(e, "INSERT INTO config (id, name, path, size, version) VALUES (?, ?, ?, ?, ?) RETURNING id",
		e.ID, e.Name, e.Path, e.Size, e.Version)

	if err != nil {
		return utils.ID{}, err
	}
	return e.ID, nil
}

func (s *ProfileStorage) Delete(id utils.ID) error {

	return s.db.Get(nil, "DELETE FROM profile WHERE id = ?", id)
}

func (s *ProfileStorage) Update(e *entity.Profile) error {

	err := s.db.Get(nil, "UPDATE profile SET name = ?, description = ?, firmware_id = ?, config_id = ? WHERE id = ?",
		e.Name, e.Description, e.FirmwareID, e.ConfigID, e.ID)
	return err
}
