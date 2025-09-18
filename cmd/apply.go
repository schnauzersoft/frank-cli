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

// Config represents the base configuration structure
type Config struct {
	Context   string `yaml:"context"`
	Namespace string `yaml:"namespace"`
}

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply [stack]",
	Short: "Apply templated Kubernetes manifest files to clusters",
	Long: `Deploy your Kubernetes applications with style and precision.

Frank reads your config files and deploys manifests to your clusters,
handling the heavy lifting of resource management and status monitoring.

What frank does:
  • Creates new resources or updates existing ones intelligently
  • Adds stack tracking annotations to keep things organized
  • Waits patiently for deployments to be ready (no more guessing!)
  • Runs multiple deployments in parallel for speed
  • Gives you clear, colored logs so you know what's happening

Target specific stacks:
  frank apply                    # Deploy everything
  frank apply dev                # Deploy all dev environment stacks
  frank apply dev/app            # Deploy all dev/app* configurations
  frank apply dev/app.yaml       # Deploy specific configuration file`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Get the --yes flag
		yes, _ := cmd.Flags().GetBool("yes")

		// Get stack filter from arguments
		var stackFilter string
		if len(args) > 0 {
			stackFilter = args[0]
		}

		// Show confirmation prompt unless --yes flag is used
		if !yes {
			if !confirmAction("apply", stackFilter) {
				fmt.Println("Canceled")
				return
			}
		}
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

		results, err := deployer.DeployAll(stackFilter)
		if err != nil {
			logger.Error("Apply failed", "error", err)
			os.Exit(1)
		}

		// Log results with appropriate log levels
		for _, result := range results {
			if result.Error != nil {
				logger.Error("Apply failed",
					"stack", result.StackName,
					"context", result.Context,
					"manifest", result.Manifest,
					"error", result.Error,
					"timestamp", result.Timestamp)
			} else {
				logger.Info("Apply successful",
					"stack", result.StackName,
					"context", result.Context,
					"manifest", result.Manifest,
					"response", result.Response,
					"timestamp", result.Timestamp)
			}
		}
	},
}

func init() {
	applyCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
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
