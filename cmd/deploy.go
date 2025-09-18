/*
Copyright © 2025 Ben Sapp ya.bsapp.ru
*/

package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"frank/pkg/deploy"

	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"
)

// Config represents the base configuration structure (context only)
type Config struct {
	Context string `yaml:"context"`
}

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy Jinja templated Kubernetes manifest files to clusters",
	Long: `Deploy command reads configuration from the config directory and deploys
the specified manifest to the Kubernetes cluster.

The config/config.yaml file should contain:
context: orbstack

Any .yaml or .yml files in the config directory (except config.yaml) will be read
as manifest config files and should contain:
manifest: sample-deployment.yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		// Set up colored structured logging
		logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{
			Level: slog.LevelDebug,
		}))

		// Find the config directory
		configDir, err := findConfigDirectory()
		if err != nil {
			logger.Error("Failed to find config directory", "error", err)
			os.Exit(1)
		}

		logger.Debug("Found config directory", "path", configDir)

		// Create deployer and run parallel deployments
		deployer := deploy.NewDeployer(configDir, logger)
		results, err := deployer.DeployAll()
		if err != nil {
			logger.Error("Deployment failed", "error", err)
			os.Exit(1)
		}

		// Log results with appropriate log levels
		for _, result := range results {
			if result.Error != nil {
				logger.Error("Deployment failed",
					"context", result.Context,
					"manifest", result.Manifest,
					"error", result.Error,
					"timestamp", result.Timestamp)
			} else {
				logger.Info("Deployment successful",
					"context", result.Context,
					"manifest", result.Manifest,
					"response", result.Response,
					"timestamp", result.Timestamp)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
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
