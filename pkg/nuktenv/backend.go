package nuktenv

import (
	"context"

	"github.com/heetch/confita/backend"
	"github.com/joho/godotenv"
)

// Backend implements Backend interface of the github.com/heetch/confita/backend
// package and provides functionality for loading and fetching configuration
// variables from a .env file.
type Backend struct {
	values map[string]string
}

// NewBackend initializes a new instance of Backend.
func NewBackend() *Backend {
	values, _ := godotenv.Read()
	return &Backend{
		values: values,
	}
}

// Get fetches a configuration variable by its key from a .env file.
func (b Backend) Get(ctx context.Context, key string) ([]byte, error) {
	if b.values == nil {
		return nil, backend.ErrNotFound
	}

	value, ok := b.values[key]
	if !ok {
		return nil, backend.ErrNotFound
	}

	return []byte(value), nil
}

// Name returns a name of this specific Backend's implementation.
func (b Backend) Name() string {
	return ".env"
}
