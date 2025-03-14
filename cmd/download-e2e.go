package cmd

import (
	"fmt"
	"os"

	"github.com/ephemeralfiles/eph/pkg/config"
	"github.com/ephemeralfiles/eph/pkg/ephcli"
	"github.com/spf13/cobra"
)

// downloadE2ECmd represents the get command
var downloadE2ECmd = &cobra.Command{
	Use:   "dle2e",
	Short: "download an object from ephemeralfiles using e2e encryption",
	Long: `download an object from ephemeralfiles using e2e encryption.
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

		err = c.DownloadE2E(uuidFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error uploading file: %s\n", err)
			os.Exit(1)
		}
	},
}
