package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var orgListFormat string

// orgListCmd represents the organization list command.
var orgListCmd = &cobra.Command{
	Use:   "list",
	Short: "List organizations",
	Long:  `List all organizations for the authenticated user.`,
	Run: func(_ *cobra.Command, _ []string) {
		InitClient()

		orgs, err := c.ListOrganizations()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing organizations: %s\n", err)
			os.Exit(1)
		}

		if len(orgs) == 0 {
			fmt.Println("No organizations found")
			return
		}

		switch orgListFormat {
		case "json":
			output, err := json.MarshalIndent(orgs, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error encoding JSON: %s\n", err)
				os.Exit(1)
			}
			fmt.Println(string(output))
		case "yaml":
			output, err := yaml.Marshal(orgs)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error encoding YAML: %s\n", err)
				os.Exit(1)
			}
			fmt.Print(string(output))
		default:
			// Table format
			tableData := pterm.TableData{
				{"ID", "NAME", "STORAGE", "SUBSCRIPTION", "ROLE"},
			}
			for _, org := range orgs {
				storage := fmt.Sprintf("%.2fGB / %.2fGB", org.UsedStorageGB, org.StorageLimitGB)
				sub := "Active"
				if !org.SubscriptionActive {
					sub = "Inactive"
				}
				role := org.UserRole
				if role == "" {
					role = "Member"
				}
				tableData = append(tableData, []string{
					org.ID,
					org.Name,
					storage,
					sub,
					role,
				})
			}
			_ = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
		}
	},
}

func init() {
	orgListCmd.Flags().StringVarP(&orgListFormat, "format", "r", "table", "output format: table, json, yaml")
}
