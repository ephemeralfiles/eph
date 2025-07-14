package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// purgeCmd represents the get command.
var purgeCmd = &cobra.Command{
	Use:   "prune",
	Short: "delete all files",
	Long: `delete all files.
`,
	Run: func(_ *cobra.Command, _ []string) {
		var gotError bool
		InitClient()
		files, err := c.Fetch()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching files: %s\n", err)
			os.Exit(1)
		}
		if files == nil {
			os.Exit(0)
		}

		for _, file := range files {
			err = c.Remove(file.FileID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error removing file %s: %s\n", file.FileID, err)
				gotError = true

				continue
			}
			fmt.Printf("File %s removed\n", file.FileID)
		}
		if gotError {
			os.Exit(1)
		}
	},
}
