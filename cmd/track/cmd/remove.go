package cmd

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/user/track/internal/config"
)

var removeCmd = &cobra.Command{
	Use:   "remove <number>",
	Short: "Remove a repository from tracking",
	Long: `Removes a repository from the tracked list by its number as shown in 'track list'.

Usage:
  track remove <number>

Aliases:
  rm

Examples:
  track remove 1
  track rm 2

Notes:
- The number refers to the index in the 'track list' table.
- This does not delete downloaded binaries or data folders (see 'track tidy' to clean up).`,
	Aliases: []string{"rm"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Get()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		num, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Printf("Error: Invalid number provided '%s'.\n", args[0])
			return
		}

		keys := make([]string, 0, len(cfg.Repos))
		for k := range cfg.Repos {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		if num < 1 || num > len(keys) {
			fmt.Printf("Error: Number %d is out of bounds.\n", num)
			return
		}

		repoToRemove := keys[num-1]
		fmt.Printf("Removing '%s' from tracking.\n", repoToRemove)
		delete(cfg.Repos, repoToRemove)

		if err := cfg.Save(); err != nil {
			fmt.Printf("Error saving configuration: %v\n", err)
		} else {
			fmt.Println("Successfully removed.")
		}
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
