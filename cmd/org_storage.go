package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ephemeralfiles/eph/pkg/ephcli"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var orgStorageFormat string

// orgStorageCmd represents the organization storage command.
var orgStorageCmd = &cobra.Command{
	Use:   "storage",
	Short: "Show organization storage information",
	Long:  `Display storage limit, usage, and availability for an organization.`,
	Run: func(_ *cobra.Command, _ []string) {
		InitClient()

		orgCtx := ephcli.NewOrgContext(c, cfg)
		org, err := orgCtx.ResolveOrganization(orgName, orgID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}

		storage, err := c.GetOrganizationStorage(org.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting storage info: %s\n", err)
			os.Exit(1)
		}

		switch orgStorageFormat {
		case "json":
			output, err := json.MarshalIndent(storage, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error encoding JSON: %s\n", err)
				os.Exit(1)
			}
			fmt.Println(string(output))
		case "yaml":
			output, err := yaml.Marshal(storage)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error encoding YAML: %s\n", err)
				os.Exit(1)
			}
			fmt.Print(string(output))
		default:
			// Table format
			fmt.Printf("Organization: %s\n", storage.Name)
			fmt.Printf("Storage Limit:    %.2f GB\n", storage.StorageLimitGB)
			fmt.Printf("Used:             %.2f GB\n", storage.UsedStorageGB)
			fmt.Printf("Available:        %.2f GB\n", storage.StorageLimitGB-storage.UsedStorageGB)
			fmt.Printf("Usage:            %.1f%%\n", storage.UsagePercent)

			status := "Normal"
			if storage.IsFull {
				status = "Full"
			} else if storage.UsagePercent > 90 {
				status = "Warning"
			}
			fmt.Printf("Status:           %s\n", status)
		}
	},
}

func init() {
	orgStorageCmd.Flags().StringVarP(&orgStorageFormat, "format", "r", "table", "output format: table, json, yaml")
}
