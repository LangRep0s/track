package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)


var rootCmd = &cobra.Command{
	Use:   "track",
	Short: "A comprehensive GitHub repository release tracker.",
	Long: `Track is a powerful CLI tool to automatically track GitHub repository releases,
download compatible binaries, and manage updates across different platforms.

You can edit the config file directly with 'track config'.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
}
