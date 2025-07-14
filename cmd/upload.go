package cmd

import (
	"github.com/ephemeralfiles/eph/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// uploadCmd represents the upload command.
var uploadCmd = &cobra.Command{
	Use:   "up",
	Short: "upload to ephemeralfiles",
	Long: `upload to ephemeralfiles.
The file is required.
`,
	Run: func(cmd *cobra.Command, _ []string) {
		InitClient()
		cmdutil.ValidateRequired(fileToUpload, "file", cmd)

		err := c.Upload(fileToUpload)
		if err != nil {
			cmdutil.HandleError("Error uploading file", err)
		}
	},
}
