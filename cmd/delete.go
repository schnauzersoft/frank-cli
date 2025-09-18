/*
Copyright © 2025 Ben Sapp ya.bsapp.ru
*/

package cmd

import (
	"github.com/spf13/cobra"
	"frank/pkg/kubernetes"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete resources managed by frank",
	Long: `Delete command removes all Kubernetes resources that have been deployed by frank.

This command finds all resources with the frankthetank.cloud/stack-name annotation
and deletes them from the current Kubernetes context.

Example usage:
  frank delete     # Delete all frank-managed resources`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get the global logger from root command
		logger := GetLogger()

		logger.Info("Starting delete process")

		// Create Kubernetes deployer for delete operations
		deployer, err := kubernetes.NewDeployerForDelete(logger)
		if err != nil {
			logger.Error("Failed to create Kubernetes deployer", "error", err)
			return
		}

		// Delete all frank-managed resources
		results, err := deployer.DeleteAllManagedResources()
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
