package dao

type Profile struct {
	Id          string `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	FirmwareId  string `db:"firmware_id"`
	ConfigId    string `db:"config_id"`
}

type Firmware struct {
	Id        string `db:"id"`
	Name      string `db:"name"`
	Path      string `db:"path"`
	Size      int64  `db:"size"`
	Version   string `db:"version"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
	UserId    string `db:"user_id"`
}

type Config struct {
	Id        string `db:"id"`
	Name      string `db:"name"`
	Path      string `db:"path"`
	Size      int64  `db:"size"`
	Version   string `db:"version"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
	UserId    string `db:"user_id"`
}
