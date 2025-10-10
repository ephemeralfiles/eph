package cmd

import (
	"github.com/spf13/cobra"
)

var (
	orgName string
	orgID   string
)

// orgCmd represents the organization command.
var orgCmd = &cobra.Command{
	Use:   "org",
	Short: "Manage organization files and settings",
	Long: `Work with organization files, storage, and settings.

Organizations allow teams to share storage and collaborate on files.
Use 'eph org use' to set a default organization context.`,
}

func init() {
	// Global flags for all org subcommands
	orgCmd.PersistentFlags().StringVar(&orgName, "org", "", "organization name")
	orgCmd.PersistentFlags().StringVar(&orgID, "org-id", "", "organization ID (UUID)")

	// Add subcommands
	orgCmd.AddCommand(orgListCmd)
	orgCmd.AddCommand(orgUseCmd)
	orgCmd.AddCommand(orgInfoCmd)
	orgCmd.AddCommand(orgStorageCmd)
	orgCmd.AddCommand(orgUploadCmd)
	orgCmd.AddCommand(orgListFilesCmd)
	orgCmd.AddCommand(orgDownloadCmd)
	orgCmd.AddCommand(orgDeleteCmd)
	orgCmd.AddCommand(orgStatsCmd)
	orgCmd.AddCommand(orgTagsCmd)

	rootCmd.AddCommand(orgCmd)
}
