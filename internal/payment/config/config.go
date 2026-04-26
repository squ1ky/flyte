package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"time"
)

type Config struct {
	Env      string `env:"ENV" env-default:"local"`
	LogLevel string `env:"PAYMENT_LOG_LEVEL" env-default:"info"`
	DB       DBConfig
	Kafka    KafkaConfig
	FakeBank FakeBankConfig
}

type DBConfig struct {
	Host     string `env:"PAYMENT_DB_HOST" env-required:"true"`
	Port     int    `env:"PAYMENT_DB_PORT" env-default:"5432"`
	User     string `env:"PAYMENT_DB_USER" env-required:"true"`
	Password string `env:"PAYMENT_DB_PASSWORD" env-required:"true"`
	Name     string `env:"PAYMENT_DB_NAME" env-required:"true"`
	SSLMode  string `env:"PAYMENT_DB_SSL_MODE" env-default:"disable"`
}

type KafkaConfig struct {
	Brokers       []string `env:"KAFKA_BROKERS" env-required:"true"`
	TopicRequests string   `env:"KAFKA_TOPIC_PAYMENT_REQUESTS" env-required:"true"`
	TopicResults  string   `env:"KAFKA_TOPIC_PAYMENT_RESULTS" env-required:"true"`
	GroupID       string   `env:"PAYMENT_KAFKA_GROUP_ID" env-required:"true"`
}

type FakeBankConfig struct {
	SuccessChancePercent int           `env:"FAKE_BANK_SUCCESS_CHANCE" env-default:"80"`
	MinDelay             time.Duration `env:"FAKE_BANK_MIN_DELAY" env-default:"500ms"`
	MaxDelay             time.Duration `env:"FAKE_BANK_MAX_DELAY" env-default:"2s"`
}

func Load() (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to read env config: %w", err)
	}

	return &cfg, nil
}
