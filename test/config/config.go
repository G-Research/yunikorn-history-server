package config

import (
	"github.com/G-Research/yunikorn-history-server/internal/config"
)

func GetTestLogConfig() *config.LogConfig {
	return &config.LogConfig{
		LogLevel: "INFO",
	}
}

func GetTestPostgresConfig() *config.PostgresConfig {
	return &config.PostgresConfig{
		Host:     "localhost",
		Port:     30002,
		Username: "postgres",
		Password: "psw",
		DbName:   "postgres",
		SSLMode:  "disable",
	}
}

func GetTestYunikornConfig() *config.YunikornConfig {
	return &config.YunikornConfig{
		Host:   "localhost",
		Port:   30001,
		Secure: false,
	}
}
