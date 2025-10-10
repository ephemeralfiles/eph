package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ephemeralfiles/eph/pkg/dto"
	"github.com/ephemeralfiles/eph/pkg/ephcli"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

const (
	bytesToKB           = 1024.0
	maxTagsDisplayLen   = 30
	maxOwnerDisplayLen  = 25
	defaultFilesLimit   = 100
	tagsTruncateLen     = 27
	ownerTruncateLen    = 22
	expirationDateLen   = 10
)

var (
	orgLsFormat  string
	orgLsTags    string
	orgLsLimit   int
	orgLsOffset  int
	orgLsRecent  bool
	orgLsExpired bool
)

// orgListFilesCmd represents the organization list files command.
var orgListFilesCmd = &cobra.Command{
	Use:   "ls",
	Short: "List organization files",
	Long:  `List files in an organization with optional filtering by tags.`,
	Run: func(_ *cobra.Command, _ []string) {
		InitClient()

		orgCtx := ephcli.NewOrgContext(c, cfg)
		org, err := orgCtx.ResolveOrganization(orgName, orgID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}

		var files []dto.OrganizationFile

		// Determine which listing method to use
		if orgLsRecent {
			files, err = c.ListRecentOrganizationFiles(org.ID, orgLsLimit)
		} else if orgLsExpired {
			files, err = c.ListExpiredOrganizationFiles(org.ID, orgLsLimit)
		} else if orgLsTags != "" {
			tags := strings.Split(orgLsTags, ",")
			for i := range tags {
				tags[i] = strings.TrimSpace(tags[i])
			}
			files, err = c.GetOrganizationFilesByTags(org.ID, tags, orgLsLimit, orgLsOffset)
		} else {
			files, err = c.ListOrganizationFiles(org.ID, orgLsLimit, orgLsOffset)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing files: %s\n", err)
			os.Exit(1)
		}

		if len(files) == 0 {
			fmt.Println("No files found")
			return
		}

		switch orgLsFormat {
		case renderFormatJSON:
			output, err := json.MarshalIndent(files, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error encoding JSON: %s\n", err)
				os.Exit(1)
			}
			fmt.Println(string(output))
		case renderFormatYAML:
			output, err := yaml.Marshal(files)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error encoding YAML: %s\n", err)
				os.Exit(1)
			}
			fmt.Print(string(output))
		case renderFormatCSV:
			fmt.Println("ID,FILENAME,SIZE,TAGS,OWNER,EXPIRATION")
			for _, file := range files {
				sizeKB := float64(file.Size) / bytesToKB
				tags := strings.Join(file.Tags, ";")
				fmt.Printf("%s,%s,%.2fKB,%s,%s,%s\n",
					file.ID, file.Filename, sizeKB, tags, file.OwnerEmail, file.ExpirationDate)
			}
		default:
			// Table format
			tableData := pterm.TableData{
				{"ID", "FILENAME", "SIZE", "TAGS", "OWNER", "EXPIRATION"},
			}
			for _, file := range files {
				sizeKB := float64(file.Size) / bytesToKB
				tags := strings.Join(file.Tags, ", ")
				if len(tags) > maxTagsDisplayLen {
					tags = tags[:tagsTruncateLen] + "..."
				}
				owner := file.OwnerEmail
				if len(owner) > maxOwnerDisplayLen {
					owner = owner[:ownerTruncateLen] + "..."
				}
				tableData = append(tableData, []string{
					file.ID,
					file.Filename,
					fmt.Sprintf("%.2fKB", sizeKB),
					tags,
					owner,
					file.ExpirationDate[:expirationDateLen],
				})
			}
			_ = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
		}
	},
}

func init() {
	orgListFilesCmd.Flags().StringVarP(&orgLsFormat, "format", "r",
		renderFormatTable, "output format: table, json, csv, yaml")
	orgListFilesCmd.Flags().StringVar(&orgLsTags, "tags", "", "filter by comma-separated tags")
	orgListFilesCmd.Flags().IntVar(&orgLsLimit, "limit", defaultFilesLimit, "maximum number of files")
	orgListFilesCmd.Flags().IntVar(&orgLsOffset, "offset", 0, "pagination offset")
	orgListFilesCmd.Flags().BoolVar(&orgLsRecent, "recent", false, "show only recent files")
	orgListFilesCmd.Flags().BoolVar(&orgLsExpired, "expired", false, "show only expired files")
}
