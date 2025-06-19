package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/track/internal/manager"
)

var (
	flagPreRelease  bool
	flagToken       string
	flagFilter      string
	flagInstallName string
)

var addCmd = &cobra.Command{
	Use:   "add <owner/repo>",
	Short: "Add a GitHub repository to track",
	Long: `Adds a new repository to the tracking list.

Examples:
  track add BurntSushi/ripgrep
  track add jesseduffield/lazygit

Flags:
  --prerelease      Include pre-releases when checking for updates
  --token           GitHub token for private repositories
  --filter          Regex to prefer a specific asset (e.g., '.*musl.*')
  --name            Set a custom binary name for the executable

After adding, an initial update is run automatically.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repoPath := args[0]
		if len(strings.Split(repoPath, "/")) != 2 {
			fmt.Println("Error: Invalid repository format. Please use 'owner/repo'.")
			return
		}

		var mgr *manager.Manager
		var err error
		if flagToken != "" {
			mgr, err = manager.NewWithToken(flagToken)
		} else {
			mgr, err = manager.New()
		}
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if err := mgr.AddRepo(repoPath); err != nil {
			fmt.Printf("Error adding repository: %v\n", err)
			return
		}

		fmt.Println("\nRunning initial update...")
		if err := mgr.UpdateRepo(repoPath, true); err != nil {
			fmt.Printf("Error during initial update: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().BoolVar(&flagPreRelease, "prerelease", false, "Include pre-releases when checking for updates")
	addCmd.Flags().StringVar(&flagToken, "token", "", "GitHub token for private repositories")
	addCmd.Flags().StringVar(&flagFilter, "filter", "", "Regex to prefer a specific asset (e.g., '.*musl.*')")
	addCmd.Flags().StringVar(&flagInstallName, "name", "", "Set a custom binary name for the executable")
}
