package cmd

import (
	"fmt"
	"os"

	"github.com/ephemeralfiles/eph/pkg/config"
	"github.com/spf13/cobra"
)

// configCmd represents the config command.
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "configure the application",
	Long: `configure the application by setting the token and endpoint.
The token is required. The endpoint is required but has a default value.
The token is the API token that you can get from the ephemeralfiles website.
The endpoint is the URL of the ephemeralfiles server.
`,
	Run: func(_ *cobra.Command, _ []string) {
		cfg := config.NewConfig()
		if token == "" {
			fmt.Fprintf(os.Stderr, "token is required\n")
			os.Exit(1)
		}
		if endpoint == "" {
			endpoint = DefaultEndpoint
		}
		cfg.Token = token
		cfg.Endpoint = endpoint

		resolvedConfigPath := config.ResolveConfigPath(configurationFile)
		err := cfg.SaveConfiguration(resolvedConfigPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error saving configuration: %s\n", err)
			os.Exit(1)
		}
		fmt.Println("Configuration saved")
	},
}
