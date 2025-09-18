/*
Copyright © 2025 Ben Sapp ya.bsapp.ru
*/

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"frank/pkg/deploy"

	"github.com/spf13/cobra"
)

// Config represents the base configuration structure (context only)
type Config struct {
	Context string `yaml:"context"`
}

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply templated Kubernetes manifest files to clusters",
	Long: `Apply command reads configuration from the config directory and applies
the specified manifest to the Kubernetes cluster.

The config/config.yaml file should contain:
context: <Kubernetes context>

Any .yaml or .yml files in the config directory (except config.yaml) will be read
as manifest config files and should contain:
manifest: example-deployment.yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get the global logger (configuration is already loaded in root command)
		logger := GetLogger()

		// Find the config directory
		configDir, err := findConfigDirectory()
		if err != nil {
			logger.Error("Failed to find config directory", "error", err)
			os.Exit(1)
		}

		logger.Debug("Found config directory", "path", configDir)

		// Create deployer and run parallel applies
		deployer, err := deploy.NewDeployer(configDir, logger)
		if err != nil {
			logger.Error("Failed to create deployer", "error", err)
			os.Exit(1)
		}

		results, err := deployer.DeployAll()
		if err != nil {
			logger.Error("Apply failed", "error", err)
			os.Exit(1)
		}

		// Log results with appropriate log levels
		for _, result := range results {
			if result.Error != nil {
				logger.Error("Apply failed",
					"context", result.Context,
					"manifest", result.Manifest,
					"error", result.Error,
					"timestamp", result.Timestamp)
			} else {
				logger.Info("Apply successful",
					"context", result.Context,
					"manifest", result.Manifest,
					"response", result.Response,
					"timestamp", result.Timestamp)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)
}

// findConfigDirectory finds the config directory by walking up the directory tree
// It only works if there's an actual 'config' directory, not just a config.yaml file
func findConfigDirectory() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting current directory: %v", err)
	}

	// First check current directory
	configPath := filepath.Join(currentDir, "config")
	if stat, err := os.Stat(configPath); err == nil && stat.IsDir() {
		// Verify it's actually a directory and has a config.yaml file
		configYamlPath := filepath.Join(configPath, "config.yaml")
		if _, err := os.Stat(configYamlPath); err == nil {
			return configPath, nil
		}
	}

	// Then check parent directory only
	parentDir := filepath.Dir(currentDir)
	if parentDir != currentDir {
		configPath := filepath.Join(parentDir, "config")
		if stat, err := os.Stat(configPath); err == nil && stat.IsDir() {
			// Verify it's actually a directory and has a config.yaml file
			configYamlPath := filepath.Join(configPath, "config.yaml")
			if _, err := os.Stat(configYamlPath); err == nil {
				return configPath, nil
			}
		}
	}

	return "", fmt.Errorf("config directory with config.yaml not found in current directory or immediate parent")
}
