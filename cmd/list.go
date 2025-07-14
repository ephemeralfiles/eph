package cmd

import (
	"fmt"
	"os"

	"github.com/ephemeralfiles/eph/pkg/ephcli"
	"github.com/spf13/cobra"
)

// listCmd represents the get command.
var listCmd = &cobra.Command{
	Use:   "ls",
	Short: "list files",
	Long: `list files. The rendering type is optional.
`,
	Run: func(_ *cobra.Command, _ []string) {
		InitClient()
		// check rendering type
		if renderingType != "table" && renderingType != "json" && renderingType != "csv" && renderingType != "yaml" {
			fmt.Fprintf(os.Stderr, "Invalid rendering type\n")
			os.Exit(1)
		}

		files, err := c.Fetch()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching files: %s\n", err)
			os.Exit(1)
		}
		if files == nil {
			os.Exit(0)
		}

		switch renderingType {
		case "json":
			err = ephcli.PrintJSON(&files)
		case "csv":
			err = ephcli.PrintCSV(&files)
		case "yaml":
			err = ephcli.PrintYAML(&files)
		default:
			err = ephcli.Print(&files)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error rendering files: %s\n", err)
			os.Exit(1)
		}
	},
}
