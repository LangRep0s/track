package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hako/durafmt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/user/track/internal/gh"
	"github.com/user/track/internal/manager"
)

var releasesCmd = &cobra.Command{
	Use:   "releases <number>",
	Short: "Show version history and recent releases for a repository",
	Long: `Shows the installed version history and recent releases from GitHub for a tracked repository.

Usage:
  track releases <number>
  track releases <number> --limit 5

Flags:
  -l, --limit   Number of recent releases to show (default 10)

Examples:
  track releases 1
  track releases 2 --limit 5

Notes:
- The number refers to the index in 'track list'.
- Shows both installed versions and recent GitHub releases.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mgr, err := manager.New()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		repoPath := getReposFromArgs([]string{args[0]}, mgr.Cfg)[0]
		if repoPath == "" {
			fmt.Println("Invalid repository number provided.")
			return
		}

		repoCfg := mgr.Cfg.Repos[repoPath]
		owner, name, _ := strings.Cut(repoPath, "/")

		fmt.Printf("Installed versions for %s (newest first):\n", repoPath)
		for _, v := range repoCfg.VersionHistory {
			fmt.Printf(" - %s\n", v)
		}
		fmt.Println()

		limit, _ := cmd.Flags().GetInt("limit")
		client := gh.NewClient(context.Background(), "")
		releases, err := client.ListReleases(context.Background(), owner, name, limit)
		if err != nil {
			fmt.Printf("Could not fetch releases from GitHub: %v\n", err)
			return
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Tag", "Name", "Published", "Type"})
		table.SetAutoWrapText(false)

		for _, rel := range releases {
			publishedAt := durafmt.ParseShort(time.Since(rel.GetPublishedAt().Time)).String()
			releaseType := "Stable"
			if rel.GetPrerelease() {
				releaseType = "Pre-release"
			}
			table.Append([]string{
				rel.GetTagName(),
				rel.GetName(),
				publishedAt + " ago",
				releaseType,
			})
		}
		fmt.Printf("Latest %d releases from GitHub:\n", limit)
		table.Render()
	},
}

func init() {
	rootCmd.AddCommand(releasesCmd)
	releasesCmd.Flags().IntP("limit", "l", 10, "Number of recent releases to show from GitHub")
}
