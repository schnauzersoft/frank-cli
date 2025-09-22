/*
Copyright © 2025 Ben Sapp ya.bsapp.ru
*/

package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	LogLevel string `mapstructure:"log_level"`
}

// LoadConfig loads configuration from environment variables and config files
// following 12-factor app principles with precedence:
// 1. Environment variables
// 2. .frank.yaml (current directory)
// 3. $HOME/.frank/config.yaml
// 4. /etc/frank/config.yaml
func LoadConfig() (*Config, error) {
	// Set up Viper
	setupViper()

	// Load config files in order of precedence
	if err := loadConfigFiles(); err != nil {
		return nil, err
	}

	// Unmarshal and normalize configuration
	return unmarshalConfig()
}

// setupViper configures Viper with environment variables and defaults
func setupViper() {
	viper.SetConfigType("yaml")
	viper.SetConfigName(".frank")
	viper.AddConfigPath(".")

	// Set environment variable prefix
	viper.SetEnvPrefix("FRANK")
	viper.AutomaticEnv()

	// Bind environment variables
	if err := viper.BindEnv("log_level", "FRANK_LOG_LEVEL"); err != nil {
		// Log the error but continue - this is not critical
		fmt.Printf("Warning: failed to bind environment variable: %v\n", err)
	}

	// Set default values
	viper.SetDefault("log_level", "info")
}

// loadConfigFiles loads configuration files in order of precedence
func loadConfigFiles() error {
	// 1. Current directory (.frank.yaml)
	if err := loadCurrentDirectoryConfig(); err != nil {
		return err
	}

	// 2. Home directory ($HOME/.frank/config.yaml)
	if err := loadHomeConfig(); err != nil {
		return err
	}

	// 3. System directory (/etc/frank/config.yaml)
	return loadSystemConfig()
}

// loadCurrentDirectoryConfig loads config from current directory
func loadCurrentDirectoryConfig() error {
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config file: %v", err)
		}
		// Config file not found is okay, we'll try other locations
	}
	return nil
}

// loadHomeConfig loads config from home directory
func loadHomeConfig() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil // Home directory not available, skip
	}

	homeConfigPath := filepath.Join(homeDir, ".frank", "config.yaml")
	if _, err := os.Stat(homeConfigPath); err != nil {
		return nil // File doesn't exist, skip
	}

	viper.SetConfigFile(homeConfigPath)
	if err := viper.MergeInConfig(); err != nil {
		return fmt.Errorf("error reading home config file: %v", err)
	}
	return nil
}

// loadSystemConfig loads config from system directory
func loadSystemConfig() error {
	systemConfigPath := "/etc/frank/config.yaml"
	if _, err := os.Stat(systemConfigPath); err != nil {
		return nil // File doesn't exist, skip
	}

	viper.SetConfigFile(systemConfigPath)
	if err := viper.MergeInConfig(); err != nil {
		return fmt.Errorf("error reading system config file: %v", err)
	}
	return nil
}

// unmarshalConfig unmarshals and normalizes the configuration
func unmarshalConfig() (*Config, error) {
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %v", err)
	}

	// Normalize log level
	config.LogLevel = strings.ToLower(config.LogLevel)

	return &config, nil
}

// GetLogLevel returns the appropriate slog.Level based on the configuration
func (c *Config) GetLogLevel() slog.Level {
	switch c.LogLevel {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// GetConfigSources returns information about which config sources were used
func GetConfigSources() []string {
	var sources []string

	// Check environment variable
	if viper.GetString("log_level") != "" {
		sources = append(sources, "environment variable FRANK_LOG_LEVEL")
	}

	// Check current directory
	if viper.ConfigFileUsed() != "" && strings.Contains(viper.ConfigFileUsed(), ".frank") {
		sources = append(sources, fmt.Sprintf("config file %s", viper.ConfigFileUsed()))
	}

	// Check home directory
	homeDir, err := os.UserHomeDir()
	if err == nil {
		homeConfigPath := filepath.Join(homeDir, ".frank", "config.yaml")
		if _, err := os.Stat(homeConfigPath); err == nil {
			sources = append(sources, fmt.Sprintf("home config file %s", homeConfigPath))
		}
	}

	// Check system directory
	systemConfigPath := "/etc/frank/config.yaml"
	if _, err := os.Stat(systemConfigPath); err == nil {
		sources = append(sources, fmt.Sprintf("system config file %s", systemConfigPath))
	}

	return sources
}
