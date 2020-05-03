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
		Host     string `config:"POSTGRES_HOST"`
		Port     string `config:"POSTGRES_PORT"`
		User     string `config:"POSTGRES_USER"`
		Password string `config:"POSTGRES_PASSWORD"`
		Name     string `config:"POSTGRES_NAME"`
		SSLMode  string `config:"POSTGRES_SSLMODE"`
	}
	RabbitMQConfig struct {
		Host     string `config:"RABBITMQ_HOST"`
		Port     string `config:"RABBITMQ_PORT"`
		Username string `config:"RABBITMQ_USERNAME"`
		Password string `config:"RABBITMQ_PASSWORD"`
	}
	ElasticSearchConfig struct {
		Host     string `config:"ELASTICSEARCH_HOST"`
		Port     string `config:"ELASTICSEARCH_PORT"`
		Username string `config:"ELASTICSEARCH_USERNAME"`
		Password string `config:"ELASTICSEARCH_PASSWORD"`
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
