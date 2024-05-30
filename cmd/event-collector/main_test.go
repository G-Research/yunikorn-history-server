package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfigFromFile(t *testing.T) {
	// Create a temporary configuration file
	tmpfile, err := os.CreateTemp("", "example.*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Write a test configuration to the temporary file
	text := []byte("yunikorn:\n  protocol: http\n  host: localhost\n  port: 8080\nyhs:\n  serverAddr: localhost:8081")
	if _, err := tmpfile.Write(text); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Set some environment variables that should be ignored
	os.Setenv("YHS_CONFIG_IGNORED", "IGNORED")

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

func TestLoadConfigFromEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("YHS_YUNIKORN_PROTOCOL", "http")
	os.Setenv("YHS_YUNIKORN_HOST", "localhost")
	os.Setenv("YHS_YUNIKORN_PORT", "8080")

	// Load with empty config file
	configFile := "this_file_does_not_exist.yaml"
	k, err := loadConfig(configFile)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "http", k.String("yunikorn.protocol"))
	assert.Equal(t, "localhost", k.String("yunikorn.host"))
	assert.Equal(t, 8080, k.Int("yunikorn.port"))
}
