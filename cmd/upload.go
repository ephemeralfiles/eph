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

By default, files are uploaded with end-to-end encryption.
Use --clear to upload without encryption.
`,
	Run: func(cmd *cobra.Command, _ []string) {
		InitClient()
		cmdutil.ValidateRequired(fileToUpload, "file", cmd)

		// Use encrypted upload by default, unless --clear flag is set
		var err error
		if clearTransfer {
			err = c.Upload(fileToUpload)
		} else {
			err = c.UploadE2E(fileToUpload)
		}

		if err != nil {
			cmdutil.HandleError("Error uploading file", err)
		}
	},
}

