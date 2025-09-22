/*
Copyright © 2025 Ben Sapp ya.bsapp.ru
*/

package cmd

import (
	"fmt"
	"os"

	"github.com/schnauzersoft/frank-cli/pkg/plan"

	"github.com/spf13/cobra"
)

// planCmd represents the plan command.
var planCmd = &cobra.Command{
	Use:   "plan [stack]",
	Short: "Show what changes would be made without applying them",
	Long: `Preview changes before applying them to your Kubernetes cluster.

frank will show you exactly what would change in your cluster, including:
  • New resources that would be created
  • Existing resources that would be updated
  • Resources that are already up to date
  • Color-coded diffs showing the changes

The output is formatted in the same format as your template files (YAML or HCL)
so you can easily see what the final manifests would look like.

Target specific stacks:
  frank plan                     # Plan all stacks
  frank plan dev                 # Plan all dev environment stacks
  frank plan dev/app             # Plan all dev/app* configurations
  frank plan dev/app.yaml        # Plan specific configuration file`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Get stack filter from arguments
		var stackFilter string
		if len(args) > 0 {
			stackFilter = args[0]
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

		// Create plan executor and run plan
		executor, err := plan.NewExecutor(configDir, logger)
		if err != nil {
			logger.Error("Failed to create plan executor", "error", err)
			os.Exit(1)
		}

		results, err := executor.PlanAll(stackFilter)
		if err != nil {
			logger.Error("Plan failed", "error", err)
			os.Exit(1)
		}

		// Display plan results
		for _, result := range results {
			if result.Error != nil {
				logger.Error("Plan failed",
					"stack", result.StackName,
					"context", result.Context,
					"manifest", result.Manifest,
					"error", result.Error)
			} else {
				fmt.Printf("\n=== Plan for %s ===\n", result.StackName)
				fmt.Printf("Context: %s\n", result.Context)
				fmt.Printf("Manifest: %s\n", result.Manifest)
				fmt.Printf("Operation: %s\n", result.Operation)
				if result.Diff != "" {
					fmt.Printf("\nDiff:\n%s\n", result.Diff)
				}
				if result.ManifestContent != "" {
					fmt.Printf("\nManifest content:\n%s\n", result.ManifestContent)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(planCmd)
}
