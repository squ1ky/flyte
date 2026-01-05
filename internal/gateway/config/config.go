package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env     string `env:"ENV" env-default:"local"`
	HTTP    HTTPConfig
	Clients ClientsConfig
}

type HTTPConfig struct {
	Port int `env:"GATEWAY_HTTP_PORT" env-default:"8080"`
}

type ClientsConfig struct {
	User ServiceConfig
}

type ServiceConfig struct {
	Addr string `env:"USER_SERVICE_ADDR" env-required:"true"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to read env config: %w", err)
	}
	return &cfg, nil
}
