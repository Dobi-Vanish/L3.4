package config

import (
	cleanenvport "github.com/wb-go/wbf/config/cleanenv-port"
)

type Config struct {
	HTTPPort     string   `env:"HTTP_PORT" env-default:"8080"`
	PostgresDSN  string   `env:"POSTGRES_DSN" env-required:"true"`
	KafkaBrokers []string `env:"KAFKA_BROKERS" env-required:"true" env-separator:","`
	KafkaTopic   string   `env:"KAFKA_TOPIC" env-default:"image-tasks"`
	StoragePath  string   `env:"STORAGE_PATH" env-default:"./storage"`
	LogLevel     string   `env:"LOG_LEVEL" env-default:"info"`
}

func Load() (*Config, error) {
	var cfg Config
	err := cleanenvport.Load(&cfg)
	return &cfg, err
}
