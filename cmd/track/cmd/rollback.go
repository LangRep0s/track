package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/user/track/internal/manager"
)

var rollbackCmd = &cobra.Command{
	Use:   "rollback <number> <version_tag>",
	Short: "Roll back a repository to a specific version",
	Long: `Downloads and installs a specific, older version of a repository.

Usage:
  track rollback <number> <version_tag>

Examples:
  track rollback 1 v1.2.3

Notes:
- The number refers to the index in 'track list'.
- The version_tag must be a valid release tag from the repository.
- This command is not implemented in this version and will print a message.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		mgr, err := manager.New()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		reposToUpdate := getReposFromArgs([]string{args[0]}, mgr.Cfg)
		if len(reposToUpdate) != 1 {
			fmt.Println("Invalid repository number provided.")
			return
		}

		fmt.Println("Rollback is not implemented in this version.")
	},
}

func init() {
	rootCmd.AddCommand(rollbackCmd)
}
