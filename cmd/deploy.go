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
		// Find the config directory
		configDir, err := findConfigDirectory()
		if err != nil {
			fmt.Printf("Error finding config directory: %v\n", err)
			os.Exit(1)
		}

		// Create deployer and run parallel deployments
		deployer := deploy.NewDeployer(configDir)
		results, err := deployer.DeployAll()
		if err != nil {
			fmt.Printf("Error during deployment: %v\n", err)
			os.Exit(1)
		}

		// Print results in the simplified format: <date> - <context> - <response>
		for _, result := range results {
			timestamp := result.Timestamp.Format("2006-01-02 15:04:05")
			if result.Error != nil {
				fmt.Printf("%s - %s - ERROR: %v\n", timestamp, result.Context, result.Error)
			} else {
				fmt.Printf("%s - %s - %s\n", timestamp, result.Context, result.Response)
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
