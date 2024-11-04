package postgres

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/G-Research/unicorn-history-server/internal/config"
)

func TestBuildConnectionInfoFromConfig(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *config.PostgresConfig
		expected string
	}{
		{
			name: "all fields populated",
			cfg: &config.PostgresConfig{
				Host:                "localhost",
				Port:                5432,
				Username:            "user1",
				Password:            "password1",
				DbName:              "dbname",
				PoolMaxConns:        10,
				PoolMinConns:        1,
				PoolMaxConnLifetime: 5 * time.Second,
				PoolMaxConnIdleTime: 10 * time.Second,
			},
			expected: "host='localhost' port='5432' user='user1' password='password1' dbname='dbname' " +
				"pool_max_conns='10' pool_min_conns='1' pool_max_conn_lifetime='5s' pool_max_conn_idle_time='10s'",
		},
		{
			name: "only required fields",
			cfg: &config.PostgresConfig{
				Host:     "localhost",
				Port:     5432,
				Username: "user2",
				Password: "password2",
				DbName:   "dbname",
			},
			expected: "host='localhost' port='5432' user='user2' password='password2' dbname='dbname'",
		},
		{
			name: "some optional fields",
			cfg: &config.PostgresConfig{
				Host:                "localhost",
				Port:                5432,
				Username:            "user",
				Password:            "password",
				DbName:              "dbname",
				PoolMaxConns:        10,
				PoolMaxConnLifetime: 7 * time.Minute,
			},
			expected: "host='localhost' port='5432' user='user' password='password' " +
				"dbname='dbname' pool_max_conns='10' pool_max_conn_lifetime='7m0s'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := buildConnectionInfoFromConfig(tt.cfg)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestBuildConnectionStringFromConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   *config.PostgresConfig
		expected string
	}{
		{
			name: "Basic connection string",
			config: &config.PostgresConfig{
				Username: "user",
				Password: "pass",
				Host:     "localhost",
				Port:     5432,
				DbName:   "mydb",
			},
			expected: "postgres://user:pass@localhost:5432/mydb",
		},
		{
			name: "With SSLMode",
			config: &config.PostgresConfig{
				Username: "user",
				Password: "pass",
				Host:     "localhost",
				Port:     5432,
				DbName:   "mydb",
				SSLMode:  "require",
			},
			expected: "postgres://user:pass@localhost:5432/mydb?sslmode=require",
		},
		{
			name: "With Schema",
			config: &config.PostgresConfig{
				Username: "user",
				Password: "pass",
				Host:     "localhost",
				Port:     5432,
				DbName:   "mydb",
				Schema:   "myschema",
			},
			expected: "postgres://user:pass@localhost:5432/mydb?search_path=myschema",
		},
		{
			name: "With SSLMode and Schema",
			config: &config.PostgresConfig{
				Username: "user",
				Password: "pass",
				Host:     "localhost",
				Port:     5432,
				DbName:   "mydb",
				SSLMode:  "require",
				Schema:   "myschema",
			},
			expected: "postgres://user:pass@localhost:5432/mydb?sslmode=require&search_path=myschema",
		},
		{
			name:     "Empty config",
			config:   &config.PostgresConfig{},
			expected: "postgres://:@:0/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildConnectionStringFromConfig(tt.config)
			if result != tt.expected {
				t.Errorf("got %s, want %s", result, tt.expected)
			}
		})
	}
}
