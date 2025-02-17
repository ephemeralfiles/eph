package cmd

import (
	"fmt"
	"os"

	"github.com/ephemeralfiles/eph/pkg/config"
	"github.com/ephemeralfiles/eph/pkg/ephcli"
	"github.com/spf13/cobra"
)

// uploadCmd represents the get command
var uploadE2ECmd = &cobra.Command{
	Use:   "upe2e",
	Short: "upload to ephemeralfiles using e2e encryption",
	Long: `upload to ephemeralfiles using e2e encryption.
The file is required.
`,
	Run: func(cmd *cobra.Command, _ []string) {
		if fileToUpload == "" {
			fmt.Fprintf(os.Stderr, "file is required\n")
			_ = cmd.Usage()
			os.Exit(1)
		}
		cfg := config.NewConfig()
		err := cfg.LoadConfiguration()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading configuration: %s\n", err)
			os.Exit(1)
		}

		c := ephcli.NewClient(cfg.Token)
		if cfg.Endpoint != "" {
			c.SetEndpoint(cfg.Endpoint)
		}
		if noProgressBar {
			c.DisableProgressBar()
		}

		err = c.UploadE2E(fileToUpload)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error uploading file: %s\n", err)
			os.Exit(1)
		}
		// fileID, pubkey, err := c.GetPublicKey()
		// if err != nil {
		// 	fmt.Fprintf(os.Stderr, "Error getting public key: %s\n", err)
		// 	os.Exit(1)
		// }
		// fmt.Println("File ID: ", fileID)
		// fmt.Println("Public Key: ", pubkey)

		// aesKey, err := GenAESKey32bits()
		// if err != nil {
		// 	fmt.Fprintf(os.Stderr, "Error generating AES key: %s\n", err)
		// 	os.Exit(1)
		// }
		// fmt.Println("AES Key: ", aesKey)
		// hexString := hex.EncodeToString(aesKey)

		// fmt.Println("aesKey: ", aesKey)
		// // convert aesKey to hexadecimal

		// // fmt.Println("encodedAESKey: ", encodedAESKey)
		// fmt.Println("hexString: ", hexString)
		// // encrypt with public key
		// encryptedAESKey, err := EncryptAESKey(pubkey, hexString)
		// if err != nil {
		// 	fmt.Fprintf(os.Stderr, "Error encrypting AES key: %s\n", err)
		// 	os.Exit(1)
		// }
		// fmt.Println("Encrypted AES Key: ", encryptedAESKey)

		// // Send the encrypted AES key to the server
		// err = c.SendAESKey(fileID, encryptedAESKey)
		// if err != nil {
		// 	fmt.Fprintf(os.Stderr, "Error sending AES key: %s\n", err)
		// 	os.Exit(1)
		// }

		// // Upload the file
		// err = c.UploadFileInChunks(aesKey, fileToUpload, c.UploadE2EEndpoint(fileID))
		// if err != nil {
		// 	fmt.Fprintf(os.Stderr, "Error uploading file: %s\n", err)
		// 	os.Exit(1)
		// }

	},
}
