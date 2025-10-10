package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/ephemeralfiles/eph/pkg/ephcli"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

const (
	defaultTagsLimit = 50
)

var (
	orgTagsFormat string
	orgTagsLimit  int
)

// orgTagsCmd represents the organization tags command.
var orgTagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "Show popular tags",
	Long:  `Display popular tags used in an organization.`,
	Run: func(_ *cobra.Command, _ []string) {
		InitClient()

		orgCtx := ephcli.NewOrgContext(c, cfg)
		org, err := orgCtx.ResolveOrganization(orgName, orgID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}

		tags, err := c.GetPopularTags(org.ID, orgTagsLimit)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting tags: %s\n", err)
			os.Exit(1)
		}

		if len(tags) == 0 {
			fmt.Println("No tags found")
			return
		}

		switch orgTagsFormat {
		case renderFormatJSON:
			output, err := json.MarshalIndent(tags, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error encoding JSON: %s\n", err)
				os.Exit(1)
			}
			fmt.Println(string(output))
		case renderFormatYAML:
			output, err := yaml.Marshal(tags)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error encoding YAML: %s\n", err)
				os.Exit(1)
			}
			fmt.Print(string(output))
		default:
			// Table format
			tableData := pterm.TableData{
				{"TAG", "COUNT"},
			}
			for _, tag := range tags {
				tableData = append(tableData, []string{
					tag.Tag,
					strconv.FormatInt(tag.Count, 10),
				})
			}
			_ = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
		}
	},
}

func init() {
	orgTagsCmd.Flags().StringVarP(&orgTagsFormat, "format", "r", renderFormatTable, "output format: table, json, yaml")
	orgTagsCmd.Flags().IntVar(&orgTagsLimit, "limit", defaultTagsLimit, "maximum number of tags (max 100)")
}
