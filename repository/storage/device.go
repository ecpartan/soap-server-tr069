package storage

import (
	"fmt"

	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/soap-server-tr069/repository/db/domain/entity"
	"github.com/ecpartan/soap-server-tr069/utils"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type DevStorage struct {
	db *sqlx.DB
}

func NewDevStorage(db *sqlx.DB) *DevStorage {
	return &DevStorage{db: db}
}

func (s *DevStorage) Get(id utils.ID) (*entity.Device, error) {
	ret := entity.Device{}

	err := s.db.Select(&ret, "SELECT * FROM device WHERE id = ?", id)

	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (s *DevStorage) GetAll(limit, offset int) []*entity.Device {
	ret := []*entity.Device{}

	err := s.db.Select(&ret, "SELECT * FROM device LIMIT $1 OFFSET $2", limit, offset)

	if err != nil {
		return nil
	}

	return ret
}

func (s *DevStorage) Search(query string) ([]*entity.Device, error) {
	var ret []*entity.Device
	var t entity.Device

	err := s.db.Get(&t, s.db.Rebind("SELECT * FROM device WHERE sn=?"), query)

	logger.LogDebug("CreateOp", err, t)

	ret = append(ret, &t)

	/*
		stmt, err := s.db.Prepare("SELECT * FROM device WHERE sn LIKE ?")

		if err != nil {
			return nil, err
		}
		defer stmt.Close()

		var t []*entity.Device
		rows, err := stmt.Query("%" + query + "%")

		for rows.Next() {
			var cur entity.Device
			err = rows.Scan(&cur.ID, &cur.SerialNumber, &cur.Mac, &cur.Username, &cur.Password, &cur.CrUsername, &cur.CrPassword, &cur.CrURL, &cur.Status, &cur.Uptime, &cur.SwVersion, &cur.HwVersion, &cur.Model, &cur.Manufacturer, &cur.OUI, &cur.Datamodel, &cur.ProfileId)
			if err != nil {
				return nil, err
			}
			t = append(t, &cur)
		}

		if len(t) == 0 {
			return nil, fmt.Errorf("not found")
		}
	*/
	return ret, nil
}

func (s *DevStorage) List() ([]*entity.Device, error) {
	rows, err := s.db.Query("SELECT * FROM device")

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var t []*entity.Device

	for rows.Next() {
		var cur entity.Device
		err = rows.Scan(&cur.ID, &cur.SerialNumber, &cur.Mac, &cur.Username, &cur.Password, &cur.CrUsername, &cur.CrPassword, &cur.CrURL, &cur.Status, &cur.Uptime, &cur.SwVersion, &cur.HwVersion, &cur.Model, &cur.Manufacturer, &cur.OUI, &cur.Datamodel, &cur.ProfileId)
		if err != nil {
			return nil, err
		}
		t = append(t, &cur)
	}

	if len(t) == 0 {
		return nil, fmt.Errorf("not found")
	}

	return t, nil
}

func (s *DevStorage) Create(e *entity.Device) (utils.ID, error) {

	err := s.db.Get(e, "INSERT INTO device (id, sn, mac, username, password, cr_username, cr_password, cr_url, status, uptime, sw_version, hw_version, model, manufacturer, oui, datamodel, profile_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		e.ID, e.SerialNumber, e.Mac, e.Username, e.Password, e.CrUsername, e.CrPassword, e.CrURL, e.Status, e.Uptime, e.SwVersion, e.HwVersion, e.Model, e.Manufacturer, e.OUI, e.Datamodel, e.ProfileId)

	if err != nil {
		return utils.ID{}, err
	}
	return e.ID, nil

}

func (s *DevStorage) Delete(id utils.ID) error {

	return s.db.Get(nil, "DELETE FROM device WHERE id = ?", id)
}

func (s *DevStorage) Update(e *entity.Device) error {

	err := s.db.Get(nil, "UPDATE device SET sn = ?, mac = ?, username = ?, password = ? WHERE id = ?", e.SerialNumber, e.Mac, e.Username, e.Password, e.ID)
	return err
}
