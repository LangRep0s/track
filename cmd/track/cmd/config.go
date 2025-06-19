package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/user/track/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Open the track configuration file for editing",
	Long: `Opens the main config file (JSON) in your default editor.

Usage:
  track config

Aliases:
  cfg

Examples:
  track config
  track cfg

Notes:
- This command opens the config file for manual editing.
- You can set pre-release, filters, and other options here.
- The config file is always stored in the global data directory.`,
	Aliases: []string{"cfg"},
	Run: func(cmd *cobra.Command, args []string) {
		_, err := config.Get()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}
		path, err := getConfigPath()
		if err != nil {
			fmt.Printf("Error finding config path: %v\n", err)
			return
		}
		if !openEditor(path) {
			fmt.Printf("You can manually edit the config file at: %s\n", path)
		}
	},
}

func getConfigPath() (string, error) {
	if runtime.GOOS == "windows" {
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			return "", fmt.Errorf("LOCALAPPDATA not set")
		}
		return localAppData + string(os.PathSeparator) + "track" + string(os.PathSeparator) + "config.json", nil
	} else {
		dataDir, err := configDataPath()
		if err != nil {
			return "", err
		}
		return dataDir + string(os.PathSeparator) + "config.json", nil
	}
}

func configDataPath() (string, error) {
	dataDir, err := os.UserCacheDir()
	if err != nil {
		home, _ := os.UserHomeDir()
		dataDir = home + string(os.PathSeparator) + ".local" + string(os.PathSeparator) + "share" + string(os.PathSeparator) + "track"
	} else {
		dataDir = dataDir + string(os.PathSeparator) + "track"
	}
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return "", err
	}
	return dataDir, nil
}

func openEditor(path string) bool {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("notepad", path)
	} else if editor := os.Getenv("EDITOR"); editor != "" {
		cmd = exec.Command(editor, path)
	} else {
		cmd = exec.Command("vi", path)
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Failed to open editor: %v\n", err)
		return false
	}
	return true
}

func init() {
	rootCmd.AddCommand(configCmd)
}
