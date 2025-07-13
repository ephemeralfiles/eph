package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// uploadCmd represents the get command
var uploadE2ECmd = &cobra.Command{
	Use:   "upe2e",
	Short: "upload to ephemeralfiles using e2e encryption",
	Long: `upload to ephemeralfiles using e2e encryption.
The file is required.
`,
	Run: func(cmd *cobra.Command, _ []string) {
		InitClient()
		if fileToUpload == "" {
			fmt.Fprintf(os.Stderr, "file is required\n")
			_ = cmd.Usage()
			os.Exit(1)
		}

		err := c.UploadE2E(fileToUpload)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error uploading file: %s\n", err)
			os.Exit(1)
		}
	},
}
