package cmd

import (
	"fmt"
	"os"

	"github.com/ephemeralfiles/eph/pkg/config"
	"github.com/ephemeralfiles/eph/pkg/ephcli"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var removeCmd = &cobra.Command{
	Use:   "rm",
	Short: "remove file from ephemeralfiles",
	Long: `remove file from ephemeralfiles.
The uuid is required.
`,
	Run: func(cmd *cobra.Command, _ []string) {
		if uuidFile == "" {
			fmt.Fprintf(os.Stderr, "uuid is required\n")
			_ = cmd.Usage()
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
		err = c.Remove(uuidFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error removing file: %s\n", err)
			os.Exit(1)
		}
	},
}
