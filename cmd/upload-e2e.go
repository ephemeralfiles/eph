package cmd

import (
	"github.com/ephemeralfiles/eph/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// uploadE2ECmd represents the upload e2e command.
var uploadE2ECmd = &cobra.Command{
	Use:   "upe2e",
	Short: "upload to ephemeralfiles using e2e encryption",
	Long: `upload to ephemeralfiles using e2e encryption.
The file is required.
`,
	Run: func(cmd *cobra.Command, _ []string) {
		InitClient()
		cmdutil.ValidateRequired(fileToUpload, "file", cmd)

		err := c.UploadE2E(fileToUpload)
		if err != nil {
			cmdutil.HandleError("Error uploading file", err)
		}
	},
}
