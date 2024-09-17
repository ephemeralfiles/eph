package cmd

import (
	"fmt"
	"os"

	"github.com/ephemeralfiles/eph/pkg/config"
	"github.com/ephemeralfiles/eph/pkg/ephcli"
	"github.com/spf13/cobra"
)

// listCmd represents the get command
var listCmd = &cobra.Command{
	Use:   "ls",
	Short: "list files",
	Long: `list files. The rendering type is optional.
`,
	Run: func(_ *cobra.Command, _ []string) {
		// check rendering type
		if renderingType != "table" && renderingType != "json" && renderingType != "csv" && renderingType != "yaml" {
			fmt.Fprintf(os.Stderr, "Invalid rendering type\n")
			os.Exit(1)
		}

		cfg := config.NewConfig()
		err := cfg.LoadConfiguration()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading configuration: %s\n", err)
			os.Exit(1)
		}

		c := ephcli.NewClient(cfg.Token)
		if cfg.Endpoint != "" {
			c.SetEndpoint(cfg.Endpoint)
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
			err = files.PrintJSON()
		case "csv":
			err = files.PrintCSV()
		case "yaml":
			err = files.PrintYAML()
		default:
			err = files.Print()
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error rendering files: %s\n", err)
			os.Exit(1)
		}
	},
}
