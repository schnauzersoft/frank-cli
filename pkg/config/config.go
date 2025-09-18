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
	viper.SetConfigType("yaml")
	viper.SetConfigName(".frank")
	viper.AddConfigPath(".")

	// Set environment variable prefix
	viper.SetEnvPrefix("FRANK")
	viper.AutomaticEnv()

	// Bind environment variables
	viper.BindEnv("log_level", "FRANK_LOG_LEVEL")

	// Set default values
	viper.SetDefault("log_level", "info")

	// Try to read config files in order of precedence
	// 1. Current directory (.frank.yaml)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %v", err)
		}
		// Config file not found is okay, we'll try other locations
	}

	// 2. Home directory ($HOME/.frank/config.yaml)
	homeDir, err := os.UserHomeDir()
	if err == nil {
		homeConfigPath := filepath.Join(homeDir, ".frank", "config.yaml")
		if _, err := os.Stat(homeConfigPath); err == nil {
			viper.SetConfigFile(homeConfigPath)
			if err := viper.MergeInConfig(); err != nil {
				return nil, fmt.Errorf("error reading home config file: %v", err)
			}
		}
	}

	// 3. System directory (/etc/frank/config.yaml)
	systemConfigPath := "/etc/frank/config.yaml"
	if _, err := os.Stat(systemConfigPath); err == nil {
		viper.SetConfigFile(systemConfigPath)
		if err := viper.MergeInConfig(); err != nil {
			return nil, fmt.Errorf("error reading system config file: %v", err)
		}
	}

	// Unmarshal configuration
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
