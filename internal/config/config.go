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
	// YHSConfig specifies the configuration for the Yunikorn History Server.
	YHSConfig YHSConfig
	// PostgresConfig specifies the configuration for the Postgres database.
	PostgresConfig PostgresConfig
	// YunikornConfig specifies the configuration for the Yunikorn API.
	YunikornConfig YunikornConfig
	// LogConfig specifies the configuration for the logger.
	LogConfig LogConfig
}

type YHSConfig struct {
	// Port specifies the port on which the Yunikorn History Server listens for incoming requests.
	Port int
	// AssetsDir specifies the directory where the static assets are stored.
	AssetsDir string
	// DataSyncInterval specifies the interval at which the data is synced from the Yunikorn API.
	DataSyncInterval time.Duration
}

func (c *YHSConfig) Validate() error {
	var errorMessages []string
	if c.Port < 1 {
		errorMessages = append(errorMessages, "yhs config validation error: port is required")
	}
	if len(errorMessages) > 0 {
		return fmt.Errorf("yhs config validation errors: %v", errorMessages)
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
// then load the configuration from environment variables prefixed with YHS_, so that environment variables take precedence.
func New(path string) (*Config, error) {
	k, err := loadConfig(path)
	if err != nil {
		return nil, err
	}

	assetsDir := k.String("yhs.assets_dir")
	if assetsDir == "" {
		assetsDir = "assets"
	}
	dataSyncInterval := k.Duration("yhs.data_sync_interval")
	if dataSyncInterval == 0 {
		dataSyncInterval = 5 * time.Minute
	}
	yhsConfig := YHSConfig{
		Port:             k.Int("yhs.port"),
		AssetsDir:        assetsDir,
		DataSyncInterval: dataSyncInterval,
	}
	if err := yhsConfig.Validate(); err != nil {
		return nil, err
	}

	yunikornConfig := YunikornConfig{
		Host:   k.String("yunikorn.host"),
		Port:   k.Int("yunikorn.port"),
		Secure: k.Bool("yunikorn.secure"),
	}

	logConfig := LogConfig{
		JSONFormat: k.Bool("log.json_format"),
		LogLevel:   k.String("log.level"),
	}

	postgresConfig := PostgresConfig{
		Host:                k.String("db.host"),
		Port:                k.Int("db.port"),
		Username:            k.String("db.user"),
		Password:            k.String("db.password"),
		DbName:              k.String("db.dbname"),
		SSLMode:             k.String("db.sslmode"),
		Schema:              k.String("db.schema"),
		PoolMaxConnLifetime: k.Duration("db.pool_max_conn_lifetime"),
		PoolMaxConnIdleTime: k.Duration("db.pool_max_conn_idletime"),
		PoolMaxConns:        k.Int("db.pool_max_conns"),
		PoolMinConns:        k.Int("db.pool_min_conns"),
	}

	config := &Config{
		YHSConfig:      yhsConfig,
		YunikornConfig: yunikornConfig,
		PostgresConfig: postgresConfig,
		LogConfig:      logConfig,
	}
	return config, nil
}

// loadConfig loads the configuration from a config file if provided,
// otherwise it loads the configuration from environment variables prefixed with YHS_.
func loadConfig(cfgFile string) (*koanf.Koanf, error) {
	k := koanf.New(".")

	if cfgFile != "" {
		if _, err := os.Stat(cfgFile); err != nil {
			return nil, fmt.Errorf("error reading config file: %v", err)
		}
		if err := k.Load(file.Provider(cfgFile), yaml.Parser()); err != nil {
			return nil, fmt.Errorf("error loading config file: %v", err)
		}
	}

	if err := k.Load(env.Provider("YHS_", ".", processEnvVar), nil); err != nil {
		return nil, fmt.Errorf("error loading environment variables: %v", err)
	}

	return k, nil
}

// Removes the prefix "YHS_" and replaces the first "_" with "."
// YHS_PARENT1_CHILD1_NAME will be converted into "parent1.child1_name"
func processEnvVar(s string) string {
	s = strings.TrimPrefix(s, "YHS_")
	firstIndex := strings.Index(s, "_")
	if firstIndex > -1 {
		s = s[:firstIndex] + "." + s[firstIndex+1:]
	}
	return strings.ToLower(s)
}
