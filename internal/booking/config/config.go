package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"time"
)

type Config struct {
	Env           string `env:"ENV" env-default:"local"`
	GRPC          GRPCConfig
	DB            DBConfig
	Kafka         KafkaConfig
	FlightService FlightServiceConfig
	Cleaner       CleanerConfig
	Outbox        OutboxConfig
}

type GRPCConfig struct {
	Port    int           `env:"BOOKING_GRPC_PORT" env-default:"50053"`
	Timeout time.Duration `env:"BOOKING_GRPC_TIMEOUT" env-default:"5s"`
}

type DBConfig struct {
	Host     string `env:"BOOKING_DB_HOST" env-required:"true"`
	Port     int    `env:"BOOKING_DB_PORT" env-default:"5432"`
	User     string `env:"BOOKING_DB_USER" env-required:"true"`
	Password string `env:"BOOKING_DB_PASSWORD" env-required:"true"`
	Name     string `env:"BOOKING_DB_NAME" env-required:"true"`
	SSLMode  string `env:"BOOKING_DB_SSL_MODE" env-default:"disable"`
}

type KafkaConfig struct {
	Brokers       []string `env:"KAFKA_BROKERS" env-default:"localhost:9092"`
	TopicRequests string   `env:"KAFKA_TOPIC_PAYMENT_REQUESTS" env-default:"payment_requests"`
	TopicResults  string   `env:"KAFKA_TOPIC_PAYMENT_RESULTS" env-default:"payment_results"`
	GroupID       string   `env:"BOOKING_KAFKA_GROUP_ID" env-default:"booking_service_group"`
}

type FlightServiceConfig struct {
	Address string `env:"FLIGHT_SERVICE_ADDR" env-required:"true"`
}

type CleanerConfig struct {
	Interval   time.Duration `env:"BOOKING_CLEANER_INTERVAL" env-default:"1m"`
	BookingTTL time.Duration `env:"RESERVATION_TTL" env-required:"true"`
}

type OutboxConfig struct {
	Interval time.Duration `env:"BOOKING_OUTBOX_INTERVAL" env-default:"5s"`
}

func Load() (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to read env config: %w", err)
	}

	return &cfg, nil
}
