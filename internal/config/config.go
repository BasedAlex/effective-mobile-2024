package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	Env *EnvSetting
}

type EnvSetting struct {
	PGDSN    string `env:"PG_DSN"`
	DBCancel int    `env:"DB_CANCEL"`
	Port     string `env:"PORT"`
	APIHost  string	`env:"API_HOST"`
	APIPort  string `env:"API_PORT"` 
	LogLevel string `env:"LOG_LEVEL"`
}

func New() *Config {
	e := &EnvSetting{}

	err := cleanenv.ReadConfig(".env", e)
	if err != nil {
		log.Panicf("read env config failed: %s", err)
	}

	return &Config{
		Env: e,
	}
}
