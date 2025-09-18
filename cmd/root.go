/*
Copyright © 2025 Ben Sapp ya.bsapp.ru
*/

package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"frank/pkg/config"

	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"
)

// Global configuration and logger
var (
	appConfig *config.Config
	logger    *slog.Logger
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "frank",
	Short: "A CLI tool for applying templated Kubernetes manifest files",
	Long: `Frank is a CLI application for applying templated Kubernetes manifest files to clusters.

It reads configuration from config/ and applies manifests from the manifests/ directory
to the specified Kubernetes cluster using the context name provided in the configuration.

Configuration for the CLI itself can be set via:
- Environment variables (FRANK_LOG_LEVEL)
- .frank.yaml (current directory)
- $HOME/.frank/config.yaml
- /etc/frank/config.yaml

Example usage:
  frank apply     # Apply manifests using config/`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Load application configuration
		var err error
		appConfig, err = config.LoadConfig()
		if err != nil {
			fmt.Printf("Failed to load configuration: %v\n", err)
			os.Exit(1)
		}

		// Set up colored structured logging with configured log level
		logger = slog.New(tint.NewHandler(os.Stdout, &tint.Options{
			Level: appConfig.GetLogLevel(),
		}))

		// Log configuration sources for debugging
		sources := config.GetConfigSources()
		if len(sources) > 0 {
			logger.Debug("Configuration loaded from", "sources", sources, "log_level", appConfig.LogLevel)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// GetConfig returns the global application configuration
func GetConfig() *config.Config {
	return appConfig
}

// GetLogger returns the global logger
func GetLogger() *slog.Logger {
	return logger
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.frank.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
}
