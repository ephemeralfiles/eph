// Package cmd provides the command-line interface commands for the eph CLI tool.
// It includes commands for file operations, configuration management, and utility functions
// for interacting with the ephemeralfiles.com service.
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/ephemeralfiles/eph/pkg/ephcli"
	"github.com/spf13/cobra"
)

// checkCmd represents the check command.
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "check configuration",
	Long: `check configuration check the current configuration and token validity.
It will display the current configuration and the box informations.
If the token is expired, it will exit with status 1.
`,
	Run: func(_ *cobra.Command, _ []string) {
		InitClient()
		email, expDate, err := ephcli.Whoami(cfg.Token)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting informations with current configuration: %s\n", err)
			os.Exit(1)
		}

		// Check if the token is expired
		if expDate.Before(time.Now()) {
			fmt.Fprintf(os.Stderr, "Token expired on %s\n", expDate.Format("2006-01-02 15:04:05"))
			os.Exit(1)
		}

		boxInfos, err := c.GetBoxInfos()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error getting box informations: %s\n", err)
			os.Exit(1)
		}

		fmt.Println("Token configuration:")
		fmt.Printf("  email: %s\n", email)
		fmt.Printf("  expiration Date: %s\n", expDate.Format("2006-01-02 15:04:05"))
		fmt.Println("Box configuration:")
		fmt.Printf("  capacity: %d MB\n", boxInfos.CapacityMb)
		fmt.Printf("  used: %d MB\n", boxInfos.UsedMb)
		fmt.Printf("  remaining: %d MB\n", boxInfos.RemainingMb)
	},
}
