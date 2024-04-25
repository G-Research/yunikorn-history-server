package config

import (
	"time"
)

type PostgresConfig struct {
	PoolMaxOpenConns    int
	PoolMaxIdleConns    int
	PoolMaxConnLifetime time.Duration
	Connection          map[string]string
}

type ECConfig struct {
	// Other settings for event-collector go here

	PostgresConfig PostgresConfig
}
