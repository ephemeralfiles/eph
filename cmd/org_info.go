package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ephemeralfiles/eph/pkg/dto"
	"github.com/ephemeralfiles/eph/pkg/ephcli"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var orgInfoFormat string

// orgInfoCmd represents the organization info command.
var orgInfoCmd = &cobra.Command{
	Use:   "info [organization-name]",
	Short: "Show organization information",
	Long:  `Display detailed information about an organization.`,
	Run: func(_ *cobra.Command, args []string) {
		InitClient()

		orgCtx := ephcli.NewOrgContext(c, cfg)

		var orgToUse string
		if len(args) > 0 {
			orgToUse = args[0]
		} else {
			orgToUse = orgName
		}

		// If no org specified via args or flag, use context resolution
		var org *dto.Organization
		var err error
		if orgToUse != "" {
			org, err = c.GetOrganizationByName(orgToUse)
			if err != nil {
				// Try as ID
				org, err = c.GetOrganization(orgToUse)
			}
		} else {
			org, err = orgCtx.ResolveOrganization(orgName, orgID)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting organization info: %s\n", err)
			os.Exit(1)
		}

		// Get storage info
		storage, err := c.GetOrganizationStorage(org.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting storage info: %s\n", err)
			os.Exit(1)
		}

		// Get stats
		stats, err := c.GetOrganizationStats(org.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting stats: %s\n", err)
			os.Exit(1)
		}

		switch orgInfoFormat {
		case renderFormatJSON:
			combined := map[string]interface{}{
				"organization": org,
				"storage":      storage,
				"stats":        stats,
			}
			output, err := json.MarshalIndent(combined, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error encoding JSON: %s\n", err)
				os.Exit(1)
			}
			fmt.Println(string(output))
		case renderFormatYAML:
			combined := map[string]interface{}{
				"organization": org,
				"storage":      storage,
				"stats":        stats,
			}
			output, err := yaml.Marshal(combined)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error encoding YAML: %s\n", err)
				os.Exit(1)
			}
			fmt.Print(string(output))
		default:
			// Table format
			fmt.Printf("Organization: %s\n", org.Name)
			fmt.Printf("ID: %s\n", org.ID)
			fmt.Printf("Storage: %.2fGB / %.2fGB (%.1f%%)\n",
				storage.UsedStorageGB, storage.StorageLimitGB, storage.UsagePercent)

			sub := "Active"
			if !org.SubscriptionActive {
				sub = "Inactive"
			}
			fmt.Printf("Subscription: %s\n", sub)
			fmt.Printf("Members: %d\n", stats.MemberCount)
			fmt.Printf("Files: %d (%d active, %d expired)\n",
				stats.FileCount, stats.ActiveFiles, stats.ExpiredFiles)
		}
	},
}

func init() {
	orgInfoCmd.Flags().StringVarP(&orgInfoFormat, "format", "r", renderFormatTable, "output format: table, json, yaml")
}
