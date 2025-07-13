package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// getCmd represents the get command.
var downloadCmd = &cobra.Command{
	Use:   "dl",
	Short: "download from ephemeralfiles",
	Long: `download from ephemeralfiles. The uuid is required.
`,
	Run: func(cmd *cobra.Command, _ []string) {
		InitClient()
		if uuidFile == "" {
			fmt.Fprintf(os.Stderr, "uuid is required\n")
			_ = cmd.Usage()
			os.Exit(1)
		}

		err := c.Download(uuidFile, "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error downloading file: %s\n", err)
			os.Exit(1)
		}
	},
}
