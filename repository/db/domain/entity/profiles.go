package entity

import (
	"time"

	"github.com/ecpartan/soap-server-tr069/utils"
)

type Profile struct {
	ID          utils.ID `db:"id"`
	Name        string   `db:"name"`
	Description string   `db:"description"`
	FirmwareID  utils.ID `db:"firmware_id"`
	ConfigID    utils.ID `db:"config_id"`
}

type ProfileView struct {
	Name        string
	Description string
}

func NewProfileView(name string, description string) ProfileView {
	return ProfileView{
		Name:        name,
		Description: description,
	}
}

func NewProfile(view ProfileView) (*Profile, error) {
	return &Profile{
		ID:          utils.NewID(),
		Name:        view.Name,
		Description: view.Description,
	}, nil
}

type Firmware struct {
	ID        utils.ID  `db:"id"`
	Name      string    `db:"name"`
	Path      string    `db:"path"`
	Size      int64     `db:"size"`
	Version   string    `db:"version"`
	CreatedAt time.Time `db:"created_at"`
}

type FirmwareView struct {
	Name    string
	Version string
	Path    string
	Size    int64
}

func NewFirmwareView(name string, version string, path string, size int64) FirmwareView {
	return FirmwareView{
		Name:    name,
		Version: version,
		Path:    path,
		Size:    size,
	}
}

type Config struct {
	ID        utils.ID  `db:"id"`
	Name      string    `db:"name"`
	Path      string    `db:"path"`
	Size      int64     `db:"size"`
	Version   string    `db:"version"`
	CreatedAt time.Time `db:"created_at"`
	UserID    utils.ID  `db:"user_id"`
}

type ConfigView struct {
	Name    string
	Version string
	Path    string
	Size    int64
}

func NewConfigView(name string, version string, path string, size int64) ConfigView {
	return ConfigView{
		Name:    name,
		Version: version,
		Path:    path,
		Size:    size,
	}
}
