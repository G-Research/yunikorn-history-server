package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testConfig = `yunikorn:
  protocol: http
  host: localhost
  port: 8080
yhs:
  serverAddr: localhost:8081
`

func TestLoadConfig_FromFileAndEnv(t *testing.T) {
	// Create a temporary configuration file
	tmpfile, err := os.CreateTemp("", "example.*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(tmpfile.Name()) })

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
	t.Cleanup(func() { os.Unsetenv("YHS_YUNIKORN_HOST") })

	// Load the configuration
	k, err := loadConfig(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "", k.String("config.ignored"))
	assert.Equal(t, "http", k.String("yunikorn.protocol"))
	assert.Equal(t, "example.com", k.String("yunikorn.host"))
	assert.Equal(t, 8080, k.Int("yunikorn.port"))
	assert.Equal(t, "localhost:8081", k.String("yhs.serverAddr"))
}

func TestLoadConfig_FromFile(t *testing.T) {
	// Create a temporary configuration file
	tmpfile, err := os.CreateTemp("", "example.*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(tmpfile.Name()) })

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

	assert.Equal(t, "", k.String("config.ignored"))
	assert.Equal(t, "http", k.String("yunikorn.protocol"))
	assert.Equal(t, "localhost", k.String("yunikorn.host"))
	assert.Equal(t, 8080, k.Int("yunikorn.port"))
	assert.Equal(t, "localhost:8081", k.String("yhs.serverAddr"))
}

func TestLoadConfig_FromEnv(t *testing.T) {
	// Set environment variables
	err := os.Setenv("YHS_YUNIKORN_PROTOCOL", "http")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Unsetenv("YHS_YUNIKORN_PROTOCOL") })
	err = os.Setenv("YHS_DB_POOL_MAX_CONN_IDLE_TIME", "120s")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Unsetenv("YHS_DB_POOL_MAX_CONN_IDLE_TIME") })

	k, err := loadConfig("")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "http", k.String("yunikorn.protocol"))
	assert.Equal(t, "120s", k.String("db.pool_max_conn_idle_time"))
}
