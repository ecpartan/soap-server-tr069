package dao

type Device struct {
	Id           string `db:"id"`
	Sn           string `db:"sn"`
	Manufacturer string `db:"manufacturer"`
	Model        string `db:"model"`
	Oui          string `db:"oui"`
	SwVersion    string `db:"sw_version"`
	HwVersion    string `db:"hw_version"`
	Ip           string `db:"ip"`
	Port         string `db:"port"`
	Uptime       int    `db:"uptime"`
	Status       string `db:"status"`
	Datamodel    string `db:"datamodel"`
	Username     string `db:"username"`
	Password     string `db:"password"`
	CrUsername   string `db:"cr_username"`
	CrPassword   string `db:"cr_password"`
	Mac          string `db:"mac"`
	LastInform   string `db:"last_inform"`
	ProfileId    string `db:"profile_id"`
}
