package cmd

import (
	"github.com/ephemeralfiles/eph/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// getCmd represents the get command.
var downloadCmd = &cobra.Command{
	Use:   "dl",
	Short: "download from ephemeralfiles",
	Long: `download from ephemeralfiles. The uuid is required.
`,
	Run: func(cmd *cobra.Command, _ []string) {
		InitClient()
		cmdutil.ValidateRequired(uuidFile, "uuid", cmd)

		err := c.Download(uuidFile, "")
		if err != nil {
			cmdutil.HandleError("Error downloading file", err)
		}
	},
}
