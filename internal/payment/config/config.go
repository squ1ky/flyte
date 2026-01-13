package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env   string `env:"ENV" env-default:"local"`
	DB    DBConfig
	Kafka KafkaConfig
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

func Load() (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to read env config: %w", err)
	}

	return &cfg, nil
}
