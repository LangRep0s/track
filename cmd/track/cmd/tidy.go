package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/track/internal/config"
)

var tidyCmd = &cobra.Command{
	Use:   "tidy",
	Short: "Delete all previous version folders for all tracked repositories (keep only current)",
	Long: `Deletes all version folders except the currently active one for each tracked repository.

Usage:
  track tidy

Examples:
  track tidy

Notes:
- This command helps free up disk space by removing old versions.
- Only the currently installed version for each repo is kept.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Get()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		for repoKey, repo := range cfg.Repos {
			if repo.CurrentVersion == "" {
				continue
			}
			_, name, _ := strings.Cut(repoKey, "/")
			repoDir := filepath.Join(cfg.Global.DataDir, name, "general")
			entries, err := os.ReadDir(repoDir)
			if err != nil {
				continue
			}
			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}
				if entry.Name() != repo.CurrentVersion {
					path := filepath.Join(repoDir, entry.Name())
					os.RemoveAll(path)
					fmt.Printf("Deleted old version: %s\n", path)
				}
			}
		}
		fmt.Println("Tidy complete.")
	},
}

func init() {
	rootCmd.AddCommand(tidyCmd)
}
