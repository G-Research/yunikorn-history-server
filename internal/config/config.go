package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"

	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	// UHSConfig specifies the configuration for the Unicorn History Server.
	UHSConfig UHSConfig
	// PostgresConfig specifies the configuration for the Postgres database.
	PostgresConfig PostgresConfig
	// YunikornConfig specifies the configuration for the Yunikorn API.
	YunikornConfig YunikornConfig
	// LogConfig specifies the configuration for the logger.
	LogConfig LogConfig
}

type UHSConfig struct {
	// Port specifies the port on which the Unicorn History Server listens for incoming requests.
	Port int
	// AssetsDir specifies the directory where the static assets are stored.
	AssetsDir string
	// DataSyncInterval specifies the interval at which the data is synced from the Yunikorn API.
	DataSyncInterval time.Duration
	// CORSConfig specifies the configuration for the CORS middleware.
	CORSConfig CORSConfig
}

type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

func (c *UHSConfig) Validate() error {
	var errorMessages []string
	if c.Port < 1 {
		errorMessages = append(errorMessages, "uhs config validation error: port is required")
	}
	if len(errorMessages) > 0 {
		return fmt.Errorf("uhs config validation errors: %v", errorMessages)
	}
	return nil
}

type PostgresConfig struct {
	Host                string
	DbName              string
	Username            string
	Password            string
	PoolMaxConnLifetime time.Duration
	PoolMaxConnIdleTime time.Duration
	Port                int
	PoolMaxConns        int
	PoolMinConns        int
	SSLMode             string
	Schema              string
}

func (c *PostgresConfig) Validate() error {
	var errorMessages []string
	if c.Host == "" {
		errorMessages = append(errorMessages, "db host is required")
	}
	if c.DbName == "" {
		errorMessages = append(errorMessages, "db name is required")
	}
	if c.Username == "" {
		errorMessages = append(errorMessages, "db user is required")
	}
	if c.Password == "" {
		errorMessages = append(errorMessages, "db password is required")
	}
	if c.Port < 1 {
		errorMessages = append(errorMessages, "db port is required")
	}
	if len(errorMessages) > 0 {
		return fmt.Errorf("postgres config validation errors: %v", errorMessages)
	}
	return nil
}

// YunikornConfig specifies the configuration for the Yunikorn API.
type YunikornConfig struct {
	Host string
	Port int
	// Secure indicates whether the connection to the Yunikorn API is using encryption or not.
	Secure bool
}

func (c *YunikornConfig) Validate() error {
	var errorMessages []string
	if c.Host == "" {
		errorMessages = append(errorMessages, "yunikorn host is required")
	}
	if c.Port < 1 {
		errorMessages = append(errorMessages, "yunikorn port is required")
	}
	if len(errorMessages) > 0 {
		return fmt.Errorf("yunikorn config validation errors: %v", errorMessages)
	}
	return nil
}

type LogConfig struct {
	LogLevel   string
	JSONFormat bool
}

// New creates a new Config object by loading the configuration from the provided path if provided,
// then load the configuration from environment variables prefixed with UHS_, so that environment variables take precedence.
func New(path string) (*Config, error) {
	k, err := loadConfig(path)
	if err != nil {
		return nil, err
	}

	assetsDir := k.String("uhs_assets_dir")
	if assetsDir == "" {
		assetsDir = "assets"
	}
	dataSyncInterval := k.Duration("uhs_data_sync_interval")
	if dataSyncInterval == 0 {
		dataSyncInterval = 5 * time.Minute
	}
	corsConfig := CORSConfig{
		AllowedOrigins: k.Strings("uhs_cors_allowed_origins"),
		AllowedMethods: k.Strings("uhs_cors_allowed_methods"),
		AllowedHeaders: k.Strings("uhs_cors_allowed_headers"),
	}

	uhsConfig := UHSConfig{
		Port:             k.Int("uhs_port"),
		AssetsDir:        assetsDir,
		DataSyncInterval: dataSyncInterval,
		CORSConfig:       corsConfig,
	}
	if err := uhsConfig.Validate(); err != nil {
		return nil, err
	}

	yunikornConfig := YunikornConfig{
		Host:   k.String("yunikorn_host"),
		Port:   k.Int("yunikorn_port"),
		Secure: k.Bool("yunikorn_secure"),
	}

	logConfig := LogConfig{
		JSONFormat: k.Bool("log_json_format"),
		LogLevel:   k.String("log_level"),
	}

	postgresConfig := PostgresConfig{
		Host:                k.String("db_host"),
		Port:                k.Int("db_port"),
		Username:            k.String("db_user"),
		Password:            k.String("db_password"),
		DbName:              k.String("db_dbname"),
		SSLMode:             k.String("db_sslmode"),
		Schema:              k.String("db_schema"),
		PoolMaxConnLifetime: k.Duration("db_pool_max_conn_lifetime"),
		PoolMaxConnIdleTime: k.Duration("db_pool_max_conn_idletime"),
		PoolMaxConns:        k.Int("db_pool_max_conns"),
		PoolMinConns:        k.Int("db_pool_min_conns"),
	}

	config := &Config{
		UHSConfig:      uhsConfig,
		YunikornConfig: yunikornConfig,
		PostgresConfig: postgresConfig,
		LogConfig:      logConfig,
	}
	return config, nil
}

// loadConfig loads the configuration from a config file if provided,
// otherwise it loads the configuration from environment variables prefixed with UHS_.
func loadConfig(cfgFile string) (*koanf.Koanf, error) {
	k := koanf.NewWithConf(koanf.Conf{
		Delim:       "_",
		StrictMerge: true,
	})

	if cfgFile != "" {
		if _, err := os.Stat(cfgFile); err != nil {
			return nil, fmt.Errorf("error reading config file: %v", err)
		}
		if err := k.Load(file.Provider(cfgFile), yaml.Parser()); err != nil {
			return nil, fmt.Errorf("error loading config file: %v", err)
		}
	}

	if err := k.Load(env.Provider("UHS_", "_", processEnvVar), nil); err != nil {
		return nil, fmt.Errorf("error loading environment variables: %v", err)
	}

	return k, nil
}

// Removes the prefix "UHS_" and converts the value to lowercase
func processEnvVar(s string) string {
	return strings.ToLower(strings.TrimPrefix(s, "UHS_"))
}
