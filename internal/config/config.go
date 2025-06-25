package config

import (
	"flag"
	"log"
	"os"
	"sync"

	dbconf "github.com/ecpartan/soap-server-tr069/db/config"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Server struct {
		Host         string `yaml:"host" json:"host" env:"SERVER_HOST"`
		Port         int    `yaml:"port" json:"port" env:"SERVER_PORT"`
		ReadTimeout  int    `yaml:"read_timeout" json:"read_timeout" env:"SERVER_READ_TIMEOUT"`
		WriteTimeout int    `yaml:"write_timeout" json:"write_timeout" env:"SERVER_WRITE_TIMEOUT"`
		IdleTimeout  int    `yaml:"idle_timeout" json:"idle_timeout" env:"SERVER_IDLE_TIMEOUT"`
		CORS         struct {
			AllowedOrigins   []string `yaml:"allowed_origins" json:"allowed_origins" env:"SERVER_CORS_ALLOWED_ORIGINS"`
			AllowedMethods   []string `yaml:"allowed_methods" json:"allowed_methods" env:"SERVER_CORS_ALLOWED_METHODS"`
			AllowedHeaders   []string `yaml:"allowed_headers" json:"allowed_headers" env:"SERVER_CORS_ALLOWED_HEADERS"`
			ExposedHeaders   []string `yaml:"exposed_headers" json:"exposed_headers" env:"SERVER_CORS_EXPOSED_HEADERS"`
			AllowCredentials bool     `yaml:"allow_credentials" json:"allow_credentials" env:"SERVER_CORS_ALLOW_CREDENTIALS"`
			MaxAge           int      `yaml:"max_age" json:"max_age" env:"SERVER_CORS_MAX_AGE"`
		} `yaml:"cors" json:"cors" env:"SERVER_CORS"`
	} `yaml:"server" json:"server" env:"SERVER"`
	dbconf.DatabaseConf `yaml:"database" json:"database" env:"DATABASE"`
	Redis               struct {
		Host           string `yaml:"host" json:"host" env:"REDIS_HOST"`
		Port           int    `yaml:"port" json:"port" env:"REDIS_PORT"`
		Password       string `yaml:"password" json:"password" env:"REDIS_PASSWORD"`
		DB             int    `yaml:"database" json:"database" env:"REDIS_DATABASE"`
		PoolSize       int    `yaml:"pool_size" json:"pool_size" env:"REDIS_POOL_SIZE"`
		MinIdleConns   int    `yaml:"min_idle_conns" json:"min_idle_conns" env:"REDIS_MIN_IDLE_CONNS"`
		MaxIdleConns   int    `yaml:"max_idle_conns" json:"max_idle_conns" env:"REDIS_MAX_IDLE_CONNS"`
		MaxActiveConns int64  `yaml:"max_active_conns" json:"max_active_conns" env:"REDIS_MAX_ACTIVE_CONNS"`
	} `yaml:"redis" json:"redis" env:"REDIS"`
}

var instance *Config
var once sync.Once
var path string

func GetConfig() *Config {
	once.Do(func() {
		flag.StringVar(&path, "config", "configs/configpgx.yaml", "Path to config file")
		flag.Parse()

		if path == "" {
			path = os.Getenv("CONFIG_PATH")
		}

		instance = &Config{}

		if err := cleanenv.ReadConfig(path, instance); err != nil {
			helptext := "Config file not found. Please create config.yaml file or set CONFIG"
			help, _ := cleanenv.GetDescription(instance, &helptext)
			log.Println(help)
			log.Fatal(err)
		}
	})
	return instance
}
