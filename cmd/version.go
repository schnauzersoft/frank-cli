/*
Copyright © 2025 Ben Sapp ya@bsapp.ru
*/

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	CommitSHA = "unknown"
	BuildTime = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long: `Display version information for the frank application.
This includes the version number, commit SHA, and build timestamp.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("frank version %s\n", Version)
		fmt.Printf("Commit: %s\n", CommitSHA)
		fmt.Printf("Built: %s\n", BuildTime)
	},
}

func GetVersionCmd() *cobra.Command {
	return versionCmd
}
