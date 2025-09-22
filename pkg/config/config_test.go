package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary directory for test config
	tempDir := t.TempDir()

	// Test with environment variable
	t.Setenv("FRANK_LOG_LEVEL", "debug")

	// Test loading config from environment
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.LogLevel != "debug" {
		t.Errorf("Expected log level 'debug', got '%s'", config.LogLevel)
	}

	// Test with config file
	configFile := filepath.Join(tempDir, ".frank.yaml")
	configContent := `log_level: info
timeout: 5m`

	err = os.WriteFile(configFile, []byte(configContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Change to temp directory to test config file loading
	t.Chdir(tempDir)

	// Clear environment variable to test file loading
	err = os.Unsetenv("FRANK_LOG_LEVEL")
	if err != nil {
		t.Errorf("Failed to unset environment variable: %v", err)
	}

	// Load config again after changing directory
	config, err = LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.LogLevel != "info" {
		t.Errorf("Expected log level 'info', got '%s'", config.LogLevel)
	}
}

func TestGetLogger(t *testing.T) {
	// Test logger creation
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	logLevel := config.GetLogLevel()
	if logLevel != slog.LevelInfo {
		t.Errorf("Expected log level to be LevelInfo, got %v", logLevel)
	}
}

func TestConfigDefaults(t *testing.T) {
	// Test default config values
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test that default values are set correctly
	if config.LogLevel == "" {
		t.Error("Expected default log level to be set")
	}
}
