package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/ephemeralfiles/eph/pkg/ephcli"
	"github.com/spf13/cobra"
)

var (
	orgUploadFile string
	orgUploadTags string
)

// orgUploadCmd represents the organization upload command.
var orgUploadCmd = &cobra.Command{
	Use:   "up",
	Short: "Upload file to organization",
	Long:  `Upload a file to an organization with optional tags.`,
	Run: func(_ *cobra.Command, _ []string) {
		InitClient()

		if orgUploadFile == "" {
			fmt.Fprintf(os.Stderr, "Error: --input flag is required\n")
			os.Exit(1)
		}

		orgCtx := ephcli.NewOrgContext(c, cfg)
		org, err := orgCtx.ResolveOrganization(orgName, orgID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}

		// Parse tags
		var tags []string
		if orgUploadTags != "" {
			tags = strings.Split(orgUploadTags, ",")
			// Trim whitespace from each tag
			for i := range tags {
				tags[i] = strings.TrimSpace(tags[i])
			}
		}

		// Upload file with E2E encryption
		fileID, err := c.UploadOrganizationFileE2E(org.ID, orgUploadFile, tags)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error uploading file: %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("File uploaded successfully with E2E encryption\n")
		fmt.Printf("File ID: %s\n", fileID)
		if len(tags) > 0 {
			fmt.Printf("Tags: %s\n", strings.Join(tags, ", "))
		}
	},
}

func init() {
	orgUploadCmd.Flags().StringVarP(&orgUploadFile, "input", "i", "", "file to upload (required)")
	orgUploadCmd.Flags().StringVar(&orgUploadTags, "tags", "", "comma-separated tags")
	orgUploadCmd.Flags().BoolVarP(&noProgressBar, "no-progress-bar", "n", false, "disable progress bar")
}
