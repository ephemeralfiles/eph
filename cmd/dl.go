package cmd

import (
	"github.com/ephemeralfiles/eph/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// downloadCmd represents the download command.
var downloadCmd = &cobra.Command{
	Use:   "dl",
	Short: "download from ephemeralfiles",
	Long: `download from ephemeralfiles. The uuid is required.

By default, files are downloaded with end-to-end encryption.
Use --clear to download without encryption.
`,
	Run: func(cmd *cobra.Command, _ []string) {
		InitClient()
		cmdutil.ValidateRequired(uuidFile, "uuid", cmd)

		// Use encrypted download by default, unless --clear flag is set
		var err error
		if clearTransfer {
			err = c.Download(uuidFile, "")
		} else {
			err = c.DownloadE2E(uuidFile)
		}
		if err != nil {
			cmdutil.HandleError("Error downloading file", err)
		}
	},
}
