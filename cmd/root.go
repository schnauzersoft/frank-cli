/*
Copyright © 2025 Ben Sapp ya.bsapp.ru
*/

package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "frank",
	Short: "A CLI tool for deploying Jinja templated Kubernetes manifest files",
	Long: `Frank is a CLI application for deploying Jinja templated Kubernetes manifest files to clusters.

It reads configuration from config/config.yaml and deploys manifests from the manifests/ directory
to the specified Kubernetes cluster using the context name provided in the configuration.

Example usage:
  frank deploy    # Deploy manifests using config/config.yaml`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.frank.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
}
