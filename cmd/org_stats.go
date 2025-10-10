package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ephemeralfiles/eph/pkg/ephcli"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var orgStatsFormat string

// orgStatsCmd represents the organization stats command.
var orgStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show organization statistics",
	Long:  `Display statistics about files, members, and usage for an organization.`,
	Run: func(_ *cobra.Command, _ []string) {
		InitClient()

		orgCtx := ephcli.NewOrgContext(c, cfg)
		org, err := orgCtx.ResolveOrganization(orgName, orgID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}

		stats, err := c.GetOrganizationStats(org.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting stats: %s\n", err)
			os.Exit(1)
		}

		switch orgStatsFormat {
		case "json":
			output, err := json.MarshalIndent(stats, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error encoding JSON: %s\n", err)
				os.Exit(1)
			}
			fmt.Println(string(output))
		case "yaml":
			output, err := yaml.Marshal(stats)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error encoding YAML: %s\n", err)
				os.Exit(1)
			}
			fmt.Print(string(output))
		default:
			// Table format
			fmt.Printf("Organization ID:  %s\n", stats.OrganizationID)
			fmt.Printf("Files:            %d\n", stats.FileCount)
			fmt.Printf("Active:           %d\n", stats.ActiveFiles)
			fmt.Printf("Expired:          %d\n", stats.ExpiredFiles)
			fmt.Printf("Total Size:       %.2f GB\n", stats.TotalSizeGB)
			fmt.Printf("Members:          %d\n", stats.MemberCount)
		}
	},
}

func init() {
	orgStatsCmd.Flags().StringVarP(&orgStatsFormat, "format", "r", "table", "output format: table, json, yaml")
}
