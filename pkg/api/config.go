package api

import (
	"context"

	"github.com/ztimes2/jazzba/pkg/nuktenv"

	"github.com/heetch/confita"
	"github.com/heetch/confita/backend/env"
)

// Config holds configuration of the service's HTTP API.
type Config struct {
	ServerPort     string `config:"SERVER_PORT"`
	PostgresConfig struct {
		Host     string `config:"DB_HOST"`
		Port     string `config:"DB_PORT"`
		User     string `config:"DB_USER"`
		Password string `config:"DB_PASSWORD"`
		Name     string `config:"DB_NAME"`
		SSLMode  string `config:"DB_SSLMODE"`
	}
	RabbitMQConfig struct {
		Host     string `config:"RMQ_HOST"`
		Port     string `config:"RMQ_PORT"`
		Username string `config:"RMQ_USERNAME"`
		Password string `config:"RMQ_PASSWORD"`
	}
	ElasticSearchConfig struct {
		Host     string `config:"ES_HOST"`
		Port     string `config:"ES_PORT"`
		Username string `config:"ES_USERNAME"`
		Password string `config:"ES_PASSWORD"`
	}
}

// LoadConfig loads configuration of the service's HTTP API.
func LoadConfig() (*Config, error) {
	var cfg Config
	if err := confita.NewLoader(
		nuktenv.NewBackend(), env.NewBackend(),
	).Load(context.Background(), &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
