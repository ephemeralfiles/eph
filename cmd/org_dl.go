package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	orgDlFile   string
	orgDlOutput string
)

// orgDownloadCmd represents the organization download command.
var orgDownloadCmd = &cobra.Command{
	Use:   "dl",
	Short: "Download file from organization",
	Long:  `Download a file from an organization by file ID.`,
	Run: func(_ *cobra.Command, _ []string) {
		InitClient()

		if orgDlFile == "" {
			fmt.Fprintf(os.Stderr, "Error: --input flag is required\n")
			os.Exit(1)
		}

		// Organization files use encrypted downloads (E2E encryption)
		// The filename is retrieved from server metadata
		err := c.DownloadE2E(orgDlFile, orgDlOutput)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error downloading file: %s\n", err)
			os.Exit(1)
		}

		fmt.Println("File downloaded successfully")
	},
}

func init() {
	orgDownloadCmd.Flags().StringVarP(&orgDlFile, "input", "i", "", "file ID to download (required)")
	orgDownloadCmd.Flags().StringVarP(&orgDlOutput, "output", "o", "", "output filename (optional)")
	orgDownloadCmd.Flags().BoolVarP(&noProgressBar, "no-progress-bar", "n", false, "disable progress bar")
}
