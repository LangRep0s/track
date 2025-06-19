package cmd

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/user/track/internal/config"
	"github.com/user/track/internal/manager"
	"github.com/user/track/internal/updater"
)

var updateCmd = &cobra.Command{
	Use:   "update [number]",
	Short: "Update tracked repositories to their latest versions (and track itself)",
	Long: `Checks for and installs new releases for all tracked repositories, or a specific one if a number is provided.

Usage:
  track update           # Update all tracked repositories and the track CLI itself
  track update 2         # Update only the repository at position 2
  track update --force   # Force update even if versions match

Examples:
  track update
  track update 1
  track update --force

Notes:
- The number refers to the index shown in 'track list'.
- After updating repositories, the track CLI will check for its own updates.
- The --force/-f flag forces an update even if the current version matches the latest.`,
	Run: func(cmd *cobra.Command, args []string) {
		mgr, err := manager.New()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			checkSelfUpdate()
			return
		}

		if len(mgr.Cfg.Repos) == 0 {
			fmt.Println("No repositories to update.")
			checkSelfUpdate()
			return
		}

		reposToUpdate := getReposFromArgs(args, mgr.Cfg)
		if len(reposToUpdate) == 0 && len(args) > 0 {
			fmt.Println("Invalid repository number provided.")
			return
		}

		forceUpdate, _ := cmd.Flags().GetBool("force")

		for _, repoPath := range reposToUpdate {
			if err := mgr.UpdateRepo(repoPath, forceUpdate); err != nil {
				fmt.Printf("Failed to update %s: %v\n", repoPath, err)
			}
			fmt.Println("---")
		}

		checkSelfUpdate()
	},
}

func checkSelfUpdate() {
	fmt.Println("Checking for updates to track CLI itself...")
	err := updater.UpdateTrack()
	if err != nil {
		fmt.Printf("track self-update failed: %v\n", err)
	} else {
		fmt.Println("track CLI was updated successfully!")
	}
}

func getReposFromArgs(args []string, cfg *config.Config) []string {
	keys := make([]string, 0, len(cfg.Repos))
	for k := range cfg.Repos {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	if len(args) == 0 {
		return keys
	}

	var repos []string
	for _, arg := range args {
		num, err := strconv.Atoi(arg)
		if err != nil || num < 1 || num > len(keys) {
			fmt.Printf("Warning: Invalid repository number '%s', skipping.\n", arg)
			continue
		}
		repos = append(repos, keys[num-1])
	}
	return repos
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().BoolP("force", "f", false, "Force update even if versions match")
}
