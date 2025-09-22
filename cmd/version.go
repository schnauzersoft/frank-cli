/*
Copyright © 2025 Ben Sapp ya@bsapp.ru
*/

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version   string
	CommitSHA string
	BuildTime string
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long: `Display version information for the frank application.
This includes the version number, commit SHA, and build timestamp.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Set defaults if not provided via ldflags
		version := Version
		if version == "" {
			version = "dev"
		}

		commit := CommitSHA
		if commit == "" {
			commit = "unknown"
		}

		buildTime := BuildTime
		if buildTime == "" {
			buildTime = "unknown"
		}

		fmt.Printf("frank version %s\n", version)
		fmt.Printf("Commit: %s\n", commit)
		fmt.Printf("Built: %s\n", buildTime)
	},
}

func GetVersionCmd() *cobra.Command {
	return versionCmd
}
