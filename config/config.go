package config

import (
	"github.com/kelseyhightower/envconfig"
)

type (
	Config struct {
		Server Server
		Mongo  Mongo
	}

	Server struct {
		Port  string `envconfig:"SERVER_PORT"`
		IsDev bool   `envconfig:"SERVER_IS_DEV"`
	}

	Mongo struct {
		Host     string `envconfig:"MONGO_HOST"`
		Port     string `envconfig:"MONGO_PORT"`
		User     string `envconfig:"MONGO_USER"`
		Password string `envconfig:"MONGO_PASSWORD"`
		Name     string `envconfig:"MONGO_NAME"`
	}
)

func New() (config Config, err error) {
	err = envconfig.Process("", &config)
	return
}
