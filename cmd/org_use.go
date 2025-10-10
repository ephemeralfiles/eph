package cmd

import (
	"fmt"
	"os"

	"github.com/ephemeralfiles/eph/pkg/config"
	"github.com/spf13/cobra"
)

var clearDefault bool

// orgUseCmd represents the organization use command.
var orgUseCmd = &cobra.Command{
	Use:   "use [organization-name]",
	Short: "Set default organization",
	Long:  `Set the default organization context for subsequent commands.`,
	Run: func(_ *cobra.Command, args []string) {
		InitClient()

		if clearDefault {
			cfg.DefaultOrganization = ""
			resolvedConfigPath := config.ResolveConfigPath(configurationFile)
			if err := cfg.SaveConfiguration(resolvedConfigPath); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving configuration: %s\n", err)
				os.Exit(1)
			}
			fmt.Println("Default organization cleared")
			return
		}

		if len(args) == 0 {
			fmt.Fprintf(os.Stderr, "Error: organization name required\n")
			fmt.Fprintf(os.Stderr, "Usage: eph org use <organization-name>\n")
			os.Exit(1)
		}

		orgName := args[0]

		// Verify organization exists
		org, err := c.GetOrganizationByName(orgName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding organization '%s': %s\n", orgName, err)
			os.Exit(1)
		}

		cfg.DefaultOrganization = org.Name
		resolvedConfigPath := config.ResolveConfigPath(configurationFile)
		if err := cfg.SaveConfiguration(resolvedConfigPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving configuration: %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("Default organization set to '%s'\n", org.Name)
	},
}

func init() {
	orgUseCmd.Flags().BoolVar(&clearDefault, "clear", false, "clear default organization")
}
