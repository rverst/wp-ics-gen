package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

var config *Config

func Get() *Config {
	if config == nil {
		config = &Config{}
		err := env.Parse(config)
		if err != nil {
			panic(err)
		}
	}
	return config
}

type Config struct {
	BaseUrl       string        `env:"BASE_URL"`
	EventsUrl     string        `env:"EVENTS_URL"`
	CheckInterval time.Duration `env:"CHECK_INTERVAL" envDefault:"5m"`
	WorkingDir    string        `env:"WORKING_DIR" envDefault:"."`
	Server        Server        `envPrefix:"SERVER_"`
	Smtp          Smtp          `envPrefix:"SMTP_"`
}

type Smtp struct {
	Host     string `env:"HOST"`
	SSL      bool   `env:"SSL" envDefault:"true"`
	Port     int    `env:"PORT" envDefault:"587"`
	Username string `env:"USERNAME"`
	Password string `env:"PASSWORD"`
}

type Server struct {
	Address string `env:"ADDRESS" envDefault:"0.0.0.0:80"`
}
