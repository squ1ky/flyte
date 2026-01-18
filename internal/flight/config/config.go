package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"time"
)

type Config struct {
	Env     string `env:"ENV" env-default:"local"`
	GRPC    GRPCConfig
	DB      DBConfig
	Elastic ElasticConfig
	Cleaner CleanerConfig
}

type GRPCConfig struct {
	Port    int           `env:"FLIGHT_GRPC_PORT" env-default:"50052"`
	Timeout time.Duration `env:"FLIGHT_GRPC_TIMEOUT" env-default:"5s"`
}

type DBConfig struct {
	Host     string `env:"FLIGHT_DB_HOST" env-required:"true"`
	Port     int    `env:"FLIGHT_DB_PORT" env-default:"5432"`
	User     string `env:"FLIGHT_DB_USER" env-required:"true"`
	Password string `env:"FLIGHT_DB_PASSWORD" env-required:"true"`
	Name     string `env:"FLIGHT_DB_NAME" env-required:"true"`
	SSLMode  string `env:"FLIGHT_DB_SSL_MODE" env-default:"disable"`
}

type ElasticConfig struct {
	URL string `env:"ELASTIC_URL" env-default:"http://localhost:9200"`
}

type CleanerConfig struct {
	Interval       time.Duration `env:"CLEANER_INTERVAL" env-default:"1m"`
	ReservationTTL time.Duration `env:"RESERVATION_TTL" env-default:"15m"`
}

func Load() (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to read env config: %w", err)
	}

	return &cfg, nil
}
