package cmd

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
	"strings"

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

		// err = c.UploadE2E(fileToUpload)
		// if err != nil {
		// 	fmt.Fprintf(os.Stderr, "Error uploading file: %s\n", err)
		// 	os.Exit(1)
		// }
		fileID, pubkey, err := c.GetPublicKey()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting public key: %s\n", err)
			os.Exit(1)
		}
		fmt.Println("File ID: ", fileID)
		fmt.Println("Public Key: ", pubkey)

		aesKey, err := GenAESKey32bits()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating AES key: %s\n", err)
			os.Exit(1)
		}
		fmt.Println("AES Key: ", aesKey)
		// encode in base64
		// encodedAESKey := base64.StdEncoding.EncodeToString(aesKey)
		hexString := hex.EncodeToString(aesKey)

		// load public key
		// rsaPubKey, err := LoadRSAPublicKey(pubkey)
		// if err != nil {
		// 	fmt.Fprintf(os.Stderr, "Error loading public key: %s\n", err)
		// 	os.Exit(1)
		// }

		fmt.Println("aesKey: ", aesKey)
		// convert aesKey to hexadecimal

		// fmt.Println("encodedAESKey: ", encodedAESKey)
		fmt.Println("hexString: ", hexString)
		// encrypt with public key
		encryptedAESKey, err := EncryptAESKey(pubkey, hexString)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error encrypting AES key: %s\n", err)
			os.Exit(1)
		}
		fmt.Println("Encrypted AES Key: ", encryptedAESKey)

		// Send the encrypted AES key to the server
		err = c.SendAESKey(fileID, encryptedAESKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error sending AES key: %s\n", err)
			os.Exit(1)
		}

		// Upload the file
		err = c.UploadFileInChunks(aesKey, fileToUpload, c.UploadE2EEndpoint(fileID))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error uploading file: %s\n", err)
			os.Exit(1)
		}

	},
}

// GenAESKey32bits generates a 32 bits AES key
func GenAESKey32bits() ([]byte, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("error generating AES key: %w", err)
	}
	return key, nil
}

func LoadRSAPublicKey(pemString string) (*rsa.PublicKey, error) {
	// Format PEM string with newlines
	formattedPEM := "-----BEGIN PUBLIC KEY-----\n"
	// Split the key into 64-character chunks
	pemString = strings.ReplaceAll(pemString, "-----BEGIN PUBLIC KEY-----", "")
	pemString = strings.ReplaceAll(pemString, "-----END PUBLIC KEY-----", "")
	pemString = strings.ReplaceAll(pemString, " ", "")

	for i := 0; i < len(pemString); i += 64 {
		end := i + 64
		if end > len(pemString) {
			end = len(pemString)
		}
		formattedPEM += pemString[i:end] + "\n"
	}
	formattedPEM += "-----END PUBLIC KEY-----"

	// Decode PEM block
	block, _ := pem.Decode([]byte(formattedPEM))
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing public key")
	}

	// Parse the public key
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing public key: %w", err)
	}

	// Cast the parsed key to *rsa.PublicKey
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}
	return rsaPub, nil
}

func formatPEM(pemString string) string {
	// Format PEM string with newlines
	formattedPEM := "-----BEGIN PUBLIC KEY-----\n"
	// Split the key into 64-character chunks
	pemString = strings.ReplaceAll(pemString, "-----BEGIN PUBLIC KEY-----", "")
	pemString = strings.ReplaceAll(pemString, "-----END PUBLIC KEY-----", "")
	pemString = strings.ReplaceAll(pemString, " ", "")

	for i := 0; i < len(pemString); i += 64 {
		end := i + 64
		if end > len(pemString) {
			end = len(pemString)
		}
		formattedPEM += pemString[i:end] + "\n"
	}
	formattedPEM += "-----END PUBLIC KEY-----"
	return formattedPEM
}

func EncryptAESKey(publicKeyPem, aesKey string) (string, error) {
	publicKeyPem = formatPEM(publicKeyPem)
	// Decode PEM block
	block, _ := pem.Decode([]byte(publicKeyPem))
	if block == nil {
		return "", fmt.Errorf("failed to parse PEM block")
	}

	// Parse public key
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse public key: %v", err)
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("not an RSA public key")
	}

	// Encrypt the message
	encryptedData, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		rsaPublicKey,
		[]byte(aesKey),
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("encryption failed: %v", err)
	}

	// Convert to base64
	return base64.StdEncoding.EncodeToString(encryptedData), nil
}
