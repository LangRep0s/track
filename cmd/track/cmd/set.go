package cmd

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/track/internal/config"
)

var setCmd = &cobra.Command{
	Use:   "set <repo#|repo> <field> <value> | set debug <true|false>",
	Short: "Set or toggle a config field for a tracked repository or global setting",
	Long: `Set or toggle a config field for a tracked repository by number (from 'track list') or by name, or set a global field like debug.

Examples:
  track set 1 prerelease true
  track set BurntSushi/ripgrep MatcherMode strict
  track set 2 AssetFilter ".*musl.*"
  track set 1 AssetPriority x86_64,amd64
  track set 2 PreferredArchives .zip,.tar.gz
  track set debug true

Supported fields:
  prerelease           (true/false)
  MatcherMode          (strict/relaxed)
  AssetFilter          (regex string)
  AssetExclude         (regex string)
  InstallName          (string)
  AssetPriority        (comma-separated list)
  PreferredArchives    (comma-separated list)
  FallbackArch         (comma-separated list)
  FallbackOS           (comma-separated list)
  debug                (true/false, global)

Use 'track list' to see repo numbers.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 2 && strings.ToLower(args[0]) == "debug" {
			return nil
		}
		if len(args) == 3 {
			return nil
		}
		return fmt.Errorf("accepts 3 arg(s) for repo fields or 2 for global debug, received %d", len(args))
	},
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Get()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		if len(args) == 2 && strings.ToLower(args[0]) == "debug" {
			debugVal := strings.ToLower(args[1])
			if debugVal == "true" {
				cfg.Global.Debug = true
			} else if debugVal == "false" {
				cfg.Global.Debug = false
			} else {
				fmt.Println("Value must be true or false for debug")
				return
			}
			if err := cfg.Save(); err != nil {
				fmt.Printf("Error saving config: %v\n", err)
			}
			return
		}
		repoKey := args[0]
		var repo *config.Repo
		if n, err := strconv.Atoi(repoKey); err == nil {
			keys := make([]string, 0, len(cfg.Repos))
			for k := range cfg.Repos {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			if n < 1 || n > len(keys) {
				fmt.Println("Invalid repo number")
				return
			}
			repo = cfg.Repos[keys[n-1]]
		} else {
			r, ok := cfg.Repos[repoKey]
			if !ok {
				fmt.Println("Repository not found")
				return
			}
			repo = r
		}
		field := strings.ToLower(args[1])
		value := args[2]
		switch field {
		case "prerelease":
			if strings.ToLower(value) == "true" {
				repo.IncludePrerelease = true
			} else if strings.ToLower(value) == "false" {
				repo.IncludePrerelease = false
			} else {
				fmt.Println("Value must be true or false")
				return
			}
		case "matchermode":
			repo.MatcherMode = strings.ToLower(value)
		case "assetfilter":
			repo.AssetFilter = value
		case "assetexclude":
			repo.AssetExclude = value
		case "installname":
			repo.InstallName = value
		case "assetpriority":
			repo.AssetPriority = strings.Split(value, ",")
		case "preferredarchives":
			repo.PreferredArchives = strings.Split(value, ",")
		case "fallbackarch":
			repo.FallbackArch = strings.Split(value, ",")
		case "fallbackos":
			repo.FallbackOS = strings.Split(value, ",")
		default:
			fmt.Println("Supported fields: prerelease, MatcherMode, AssetFilter, AssetExclude, InstallName, AssetPriority, PreferredArchives, FallbackArch, FallbackOS, debug (global)")
			return
		}
		if err := cfg.Save(); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(setCmd)
}
