package entity

import (
	"net"
	"time"

	"github.com/ecpartan/soap-server-tr069/utils"
)

type Device struct {
	ID           utils.ID  `db:"id"`
	SerialNumber string    `db:"sn"`
	Manufacturer string    `db:"manufacturer"`
	Model        string    `db:"model"`
	OUI          string    `db:"oui"`
	SwVersion    string    `db:"sw_version"`
	HwVersion    string    `db:"hw_version"`
	Uptime       int       `db:"uptime"`
	Status       string    `db:"status"`
	Datamodel    string    `db:"datamodel"`
	Username     string    `db:"username"`
	Password     string    `db:"password"`
	CrUsername   string    `db:"cr_username"`
	CrPassword   string    `db:"cr_password"`
	CrURL        string    `db:"cr_url"`
	Mac          string    `db:"mac"`
	Created_at   time.Time `db:"created_at"`
	Updated_at   time.Time `db:"updated_at"`
	ProfileId    utils.ID  `db:"profile_id"`
}

type DeviceView struct {
	SerialNumber string
	Manufacturer string
	Model        string
	OUI          string
	SwVersion    string
	HwVersion    string
	Datamodel    string
	CrURL        string
}

type DeviceAuthView struct {
	Username   string
	Password   string
	CrUsername string
	CrPassword string
}

func NewDeviceView(serial, man, model, oui, sw, hw, datamodel, crurl string) *DeviceView {
	return &DeviceView{
		SerialNumber: serial,
		Manufacturer: man,
		Model:        model,
		OUI:          oui,
		SwVersion:    sw,
		HwVersion:    hw,
		Datamodel:    datamodel,
		CrURL:        crurl,
	}
}

func NewDeviceAuthView(username, password, crusername, crpassword string) *DeviceAuthView {
	return &DeviceAuthView{
		Username:   username,
		Password:   password,
		CrUsername: crusername,
		CrPassword: crpassword,
	}
}

func NewDevice(view DeviceView) (*Device, error) {
	b := &Device{
		ID:           utils.NewID(),
		SerialNumber: view.SerialNumber,
		Manufacturer: view.Manufacturer,
		Model:        view.Model,
		Created_at:   time.Now(),
		Updated_at:   time.Now(),
		OUI:          view.OUI,
		SwVersion:    view.SwVersion,
		HwVersion:    view.HwVersion,
		Datamodel:    view.Datamodel,
		CrURL:        view.CrURL,
		CrUsername:   "",
		CrPassword:   "",
		Username:     "",
		Password:     "",
		Status:       "off",
		Uptime:       0,
	}
	err := b.Validate()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (d *Device) Validate() error {
	if len(d.SerialNumber) < 1 || len(d.SerialNumber) > 255 {
		return ErrInvalidEntity
	}
	if len(d.OUI) < 1 || len(d.OUI) > 20 {
		return ErrInvalidEntity
	}
	if len(d.Manufacturer) < 1 || len(d.Manufacturer) > 20 {
		return ErrInvalidEntity
	}
	if net.ParseIP(d.CrURL) == nil {
		return ErrCRULR
	}
	return nil
}
