package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "development"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print version of application",
	Long:  `print version of application`,
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println(version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
