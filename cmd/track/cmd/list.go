package cmd

import (
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/user/track/internal/config"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tracked repositories",
	Long: `Lists all repositories currently tracked by track, showing their current version, pre-release status, and asset filter (if any).

Usage:
  track list

Aliases:
  ls, status

Examples:
  track list
  track ls

This command displays a table of all tracked repositories. If none are tracked, it will prompt you to add one.`,
	Aliases: []string{"ls", "status"},
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Get()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if len(cfg.Repos) == 0 {
			fmt.Println("No repositories are being tracked. Use 'track add <owner/repo>' to add one.")
			return
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"#", "Repository", "Current Version", "Pre-release", "Filter"})
		table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
		table.SetCenterSeparator("|")

		keys := make([]string, 0, len(cfg.Repos))
		for k := range cfg.Repos {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for i, k := range keys {
			repo := cfg.Repos[k]
			version := repo.CurrentVersion
			if version == "" {
				version = "Not installed"
			}
			table.Append([]string{
				strconv.Itoa(i + 1),
				k,
				version,
				fmt.Sprintf("%t", repo.IncludePrerelease),
				repo.AssetFilter,
			})
		}
		table.Render()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
