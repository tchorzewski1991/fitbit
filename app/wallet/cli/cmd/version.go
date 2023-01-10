package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var version = "0.0.1"

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints current version",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("Version: %s\n", version)
	},
}
