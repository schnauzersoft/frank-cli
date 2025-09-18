/*
Copyright © 2025 Ben Sapp ya.bsapp.ru
*/

package cmd

import (
	"frank/pkg/kubernetes"

	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete [stack]",
	Short: "Delete resources managed by frank",
	Long: `Clean up your Kubernetes resources with surgical precision.

Frank finds and removes resources it previously deployed, using stack
annotations to identify what belongs to frank vs other tools.

What frank deletes:
  • Only resources with frankthetank.cloud/stack-name annotations
  • Matches by stack name patterns for selective cleanup
  • Searches across all namespaces to find everything
  • Shows you exactly what it's removing with clear logs

Target specific stacks:
  frank delete                    # Remove all frank-managed resources
  frank delete dev                # Remove all dev environment resources
  frank delete dev/app            # Remove all dev/app* stack resources
  frank delete frank-dev-app      # Remove specific stack`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Get the global logger from root command
		logger := GetLogger()

		// Get stack filter from arguments
		var stackFilter string
		if len(args) > 0 {
			stackFilter = args[0]
		}

		logger.Info("Starting delete process", "filter", stackFilter)

		// Create Kubernetes deployer for delete operations
		deployer, err := kubernetes.NewDeployerForDelete(logger)
		if err != nil {
			logger.Error("Failed to create Kubernetes deployer", "error", err)
			return
		}

		// Delete frank-managed resources (with optional filtering)
		results, err := deployer.DeleteAllManagedResources(stackFilter)
		if err != nil {
			logger.Error("Delete process failed", "error", err)
			return
		}

		// Log results
		for _, result := range results {
			if result.Error != nil {
				logger.Error("Delete failed",
					"stack", result.StackName,
					"resource", result.ResourceType,
					"name", result.ResourceName,
					"namespace", result.Namespace,
					"error", result.Error)
			} else {
				logger.Info("Delete successful",
					"stack", result.StackName,
					"resource", result.ResourceType,
					"name", result.ResourceName,
					"namespace", result.Namespace)
			}
		}

		logger.Info("Delete process completed", "total_resources", len(results))
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
