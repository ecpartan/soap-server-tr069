package config

type DatabaseConf struct {
	Host     string `yaml:"host" json:"host" env:"DATABASE_HOST"`
	Port     int    `yaml:"port" json:"port" env:"DATABASE_PORT"`
	UserName string `yaml:"username" json:"username" env:"DATABASE_USERNAME"`
	Password string `yaml:"password" json:"password" env:"DATABASE_PASSWORD"`
	Database string `yaml:"database" json:"database" env:"DATABASE_DATABASE"`
	Driver   string `yaml:"driver" json:"driver" env:"DATABASE_DRIVER"`
}
