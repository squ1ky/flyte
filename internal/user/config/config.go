package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"time"
)

type Config struct {
	Env  string `env:"ENV" env-default:"local"`
	GRPC GRPCConfig
	DB   DBConfig
	JWT  JWTConfig
}

type GRPCConfig struct {
	Port    int           `env:"USER_GRPC_PORT" env-default:"50051"`
	Timeout time.Duration `env:"USER_GRPC_TIMEOUT" env-default:"5s"`
}

type DBConfig struct {
	Host     string `env:"USER_DB_HOST" env-required:"true"`
	Port     int    `env:"USER_DB_PORT" env-default:"5432"`
	User     string `env:"USER_DB_USER" env-required:"true"`
	Password string `env:"USER_DB_PASSWORD" env-required:"true"`
	Name     string `env:"USER_DB_NAME" env-required:"true"`
	SSLMode  string `env:"USER_DB_SSL_MODE" env-default:"disable"`
}

type JWTConfig struct {
	Secret string        `env:"USER_JWT_SECRET" env-required:"true"`
	TTL    time.Duration `env:"USER_JWT_TTL" env-default:"24h"`
}

func Load() (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to read env config: %w", err)
	}

	return &cfg, nil
}
