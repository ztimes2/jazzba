package postgres

import (
	"database/sql"
	"fmt"

	"github.com/Masterminds/squirrel"
	// Loads PostgreSQL's driver as a side effect.
	_ "github.com/lib/pq"
)

const (
	driverName = "postgres"
	dsnFormat  = "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s"
)

// SSLMode represents an SSL mode of PostgreSQL.
type SSLMode string

const (
	// DisableSSLMode is used when you don't care about security, and you don't
	// want to pay the overhead of encryption.
	DisableSSLMode SSLMode = "disable"
)

// Config holds configuration for connecting to PostgreSQL.
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  SSLMode
}

// NewDB opens a new PostgreSQL database connection.
func NewDB(cfg Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		dsnFormat,
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.SSLMode,
	)

	return sql.Open(driverName, dsn)
}

var sqlQueryBase = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
