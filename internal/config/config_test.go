package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/rs/cors"
	"github.com/stretchr/testify/assert"
)

const testConfig = `yunikorn:
  protocol: http
  host: localhost
  port: 8080
yhs:
  serverAddr: localhost:8081
`

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		want    *Config
		wantErr bool
	}{
		{
			name: "valid config file",
			path: filepath.Join("testdata", "config.yml"),
			want: &Config{
				YHSConfig: YHSConfig{
					Port:             8080,
					AssetsDir:        "assets",
					DataSyncInterval: 5 * time.Minute,
					CORSConfig: cors.Options{
						AllowedOrigins: []string{"*"},
						AllowedMethods: []string{"GET"},
						AllowedHeaders: []string{"*"},
					},
				},
				YunikornConfig: YunikornConfig{
					Host:   "localhost",
					Port:   9090,
					Secure: false,
				},
				LogConfig: LogConfig{
					LogLevel:   "info",
					JSONFormat: false,
				},
				PostgresConfig: PostgresConfig{
					Host:                "localhost",
					DbName:              "testdb",
					Username:            "user",
					Password:            "password",
					Port:                5432,
					PoolMaxConnLifetime: 30 * time.Minute,
					PoolMaxConnIdleTime: 10 * time.Minute,
					PoolMaxConns:        10,
					PoolMinConns:        1,
					SSLMode:             "disable",
				},
			},
			wantErr: false,
		},
		{
			name:    "missing config file",
			path:    filepath.Join("testdata", "missing_config.yml"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NoError(t, os.Setenv("YHS_DB_PASSWORD", "password"))
			t.Cleanup(func() { assert.NoError(t, os.Unsetenv("YHS_DB_PASSWORD")) })
			got, err := New(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			diff := cmp.Diff(tt.want, got)
			if diff != "" {
				t.Errorf("New() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestYHSConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  YHSConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: YHSConfig{
				Port: 8080,
			},
			wantErr: false,
		},
		{
			name: "invalid config - port missing",
			config: YHSConfig{
				Port: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("YHSConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPostgresConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  PostgresConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: PostgresConfig{
				Host:     "localhost",
				DbName:   "testdb",
				Username: "user",
				Password: "password",
				Port:     5432,
			},
			wantErr: false,
		},
		{
			name: "invalid config - missing host",
			config: PostgresConfig{
				DbName:   "testdb",
				Username: "user",
				Password: "password",
				Port:     5432,
			},
			wantErr: true,
		},
		{
			name: "invalid config - missing db name",
			config: PostgresConfig{
				Host:     "localhost",
				Username: "user",
				Password: "password",
				Port:     5432,
			},
			wantErr: true,
		},
		{
			name: "invalid config - missing username",
			config: PostgresConfig{
				Host:     "localhost",
				DbName:   "testdb",
				Password: "password",
				Port:     5432,
			},
			wantErr: true,
		},
		{
			name: "invalid config - missing password",
			config: PostgresConfig{
				Host:     "localhost",
				DbName:   "testdb",
				Username: "user",
				Port:     5432,
			},
			wantErr: true,
		},
		{
			name: "invalid config - missing port",
			config: PostgresConfig{
				Host:     "localhost",
				DbName:   "testdb",
				Username: "user",
				Password: "password",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("PostgresConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestYunikornConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  YunikornConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: YunikornConfig{
				Host: "localhost",
				Port: 8080,
			},
			wantErr: false,
		},
		{
			name: "invalid config - missing host",
			config: YunikornConfig{
				Port: 8080,
			},
			wantErr: true,
		},
		{
			name: "invalid config - missing port",
			config: YunikornConfig{
				Host: "localhost",
				Port: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("YunikornConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadConfig_FromFileAndEnv(t *testing.T) {
	// Create a temporary configuration file
	tmpfile, err := os.CreateTemp("", "example.*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(tmpfile.Name()) })

	// Write a test configuration to the temporary file
	text := []byte(testConfig)
	if _, err := tmpfile.Write(text); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Set environment variables
	if err = os.Setenv("YHS_YUNIKORN_HOST", "example.com"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("YHS_YUNIKORN_HOST") })

	// Load the configuration
	k, err := loadConfig(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "", k.String("config_ignored"))
	assert.Equal(t, "http", k.String("yunikorn_protocol"))
	assert.Equal(t, "example.com", k.String("yunikorn_host"))
	assert.Equal(t, 8080, k.Int("yunikorn_port"))
	assert.Equal(t, "localhost:8081", k.String("yhs_serverAddr"))
}

func TestLoadConfig_FromFile(t *testing.T) {
	// Create a temporary configuration file
	tmpfile, err := os.CreateTemp("", "example.*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(tmpfile.Name()) })

	// Write a test configuration to the temporary file
	text := []byte(testConfig)
	if _, err := tmpfile.Write(text); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	k, err := loadConfig(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "", k.String("config_ignored"))
	assert.Equal(t, "http", k.String("yunikorn_protocol"))
	assert.Equal(t, "localhost", k.String("yunikorn_host"))
	assert.Equal(t, 8080, k.Int("yunikorn_port"))
	assert.Equal(t, "localhost:8081", k.String("yhs_serverAddr"))
}

func TestLoadConfig_FromEnv(t *testing.T) {
	// Set environment variables
	err := os.Setenv("YHS_DB_PASSWORD", "psw")
	if err != nil {
		t.Fatal(err)
	}
	err = os.Setenv("YHS_YUNIKORN_PROTOCOL", "http")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("YHS_YUNIKORN_PROTOCOL") })
	err = os.Setenv("YHS_DB_POOL_MAX_CONN_IDLE_TIME", "120s")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("YHS_DB_POOL_MAX_CONN_IDLE_TIME") })

	k, err := loadConfig("")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "http", k.String("yunikorn_protocol"))
	assert.Equal(t, "120s", k.String("db_pool_max_conn_idle_time"))
	assert.Equal(t, "psw", k.String("db_password"))
}
