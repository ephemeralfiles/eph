package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// uploadCmd represents the get command
var uploadCmd = &cobra.Command{
	Use:   "up",
	Short: "upload to ephemeralfiles",
	Long: `upload to ephemeralfiles.
The file is required.
`,
	Run: func(cmd *cobra.Command, _ []string) {
		InitClient()
		if fileToUpload == "" {
			fmt.Fprintf(os.Stderr, "file is required\n")
			_ = cmd.Usage()
			os.Exit(1)
		}

		err := c.Upload(fileToUpload)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error uploading file: %s\n", err)
			os.Exit(1)
		}
	},
}
