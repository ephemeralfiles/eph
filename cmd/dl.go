package cmd

import (
	"fmt"
	"os"

	"github.com/ephemeralfiles/eph/pkg/config"
	"github.com/ephemeralfiles/eph/pkg/ephcli"
	"github.com/spf13/cobra"
)

// getCmd represents the get command.
var downloadCmd = &cobra.Command{
	Use:   "dl",
	Short: "download from ephemeralfiles",
	Long: `download from ephemeralfiles. The uuid is required.
`,
	Run: func(cmd *cobra.Command, _ []string) {
		if uuidFile == "" {
			fmt.Fprintf(os.Stderr, "uuid is required\n")
			_ = cmd.Usage()
			os.Exit(1)
		}

		cfg := config.NewConfig()
		err := cfg.LoadConfiguration(configurationFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading configuration: %s\n", err)
			os.Exit(1)
		}

		c := ephcli.NewClient(cfg.Token)
		if cfg.Endpoint != "" {
			c.SetEndpoint(cfg.Endpoint)
		}
		if noProgressBar {
			c.DisableProgressBar()
		}

		err = c.Download(uuidFile, "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error downloading file: %s\n", err)
			os.Exit(1)
		}
	},
}
