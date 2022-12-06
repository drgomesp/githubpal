package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version string
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "A version of software",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Printf("version: %s\n", Version)

	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
