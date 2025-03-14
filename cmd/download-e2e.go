package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// downloadE2ECmd represents the get command
var downloadE2ECmd = &cobra.Command{
	Use:   "dle2e",
	Short: "download an object from ephemeralfiles using e2e encryption",
	Long: `download an object from ephemeralfiles using e2e encryption.
`,
	Run: func(cmd *cobra.Command, _ []string) {
		InitClient()
		if uuidFile == "" {
			fmt.Fprintf(os.Stderr, "uuid is required\n")
			_ = cmd.Usage()
			os.Exit(1)
		}

		err := c.DownloadE2E(uuidFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error uploading file: %s\n", err)
			os.Exit(1)
		}
	},
}
