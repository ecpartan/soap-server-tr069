package dto



type CreateDeviceInput struct {
	Sn           string
	Manufacturer string
	Model        string
	Oui          string
	SwVersion    string
	HwVersion    string
	Ip           string
	Port         string
	Uptime       int
	Status       string
	Datamodel    string
	Username     string
	Password     string
	CrUsername   string
	CrPassword   string
	Mac          string
	LastInform   string
	ProfileId    string
}

