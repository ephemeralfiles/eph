package ephcli

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/ephemeralfiles/eph/pkg/dto"
	"github.com/schollz/progressbar/v3"
)

func (c *ClientEphemeralfiles) GetFileInformationEndpoint(fileID string) string {
	return fmt.Sprintf("%s/%s/files/info/%s", c.endpoint, apiVersion, fileID)
}

func (c *ClientEphemeralfiles) GetNewDownloadTransactionEndpoint(fileID string) string {
	return fmt.Sprintf("%s/%s/public-key/download/%s", c.endpoint, apiVersion, fileID)
}

func (c *ClientEphemeralfiles) CreateNewDownloadTransaction(fileID string) (transactionID, publicKey string, err error) {
	req, err := http.NewRequest(http.MethodHead, c.GetNewDownloadTransactionEndpoint(fileID), nil)
	if err != nil {
		return "", "", fmt.Errorf("error creating request: %w", err)
	}
	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Get Header X-File-Id from Header
	transactionID = resp.Header.Get("X-Transaction-Id")
	if transactionID == "" {
		return "", "", fmt.Errorf("error reading response: %w", err)
	}
	// Get Header X-File-Public-Key from Header
	publicKey = resp.Header.Get("X-File-Public-Key")
	if publicKey == "" {
		return "", "", fmt.Errorf("error reading response: %w", err)
	}
	return transactionID, publicKey, nil
}

func (c *ClientEphemeralfiles) DownloadE2E(fileID string) error {
	// Get file information
	req, err := http.NewRequest(http.MethodGet, c.GetFileInformationEndpoint(fileID), nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Decode response
	var fileInfo dto.InfoFile
	err = json.NewDecoder(resp.Body).Decode(&fileInfo)
	if err != nil {
		return fmt.Errorf("error decoding response: %w", err)
	}

	fmt.Println("File Information: ", fileInfo)

	transactionID, pubkey, err := c.CreateNewDownloadTransaction(fileID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting public key: %s\n", err)
		os.Exit(1)
	}
	fmt.Println("File ID: ", transactionID)
	fmt.Println("Public Key: ", pubkey)

	aesKey, err := GenAESKey32bits()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating AES key: %s\n", err)
		os.Exit(1)
	}
	fmt.Println("AES Key: ", aesKey)
	hexString := hex.EncodeToString(aesKey)

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
	err = c.SendAESKeyForDownloadTransaction(transactionID, encryptedAESKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error sending AES key: %s\n", err)
		os.Exit(1)
	}

	// Create progress bar
	bar := progressbar.NewOptions64(fileInfo.Size, progressbar.OptionClearOnFinish(),
		progressbar.OptionShowBytes(true), progressbar.OptionSetWidth(DefaultBarWidth),
		progressbar.OptionSetDescription("downloading file..."),
		progressbar.OptionSetVisibility(!c.noProgressBar),
	)

	for i := 0; i < fileInfo.NbParts; i++ {
		// fmt.Println("Downloading part ", i)
		chunkSize, err := c.DownloadPartE2E(fileInfo.Filename, transactionID, aesKey, i)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error downloading part: %s\n", err)
			os.Exit(1)
		}
		_ = bar.Add(chunkSize)
	}
	bar.Clear()
	bar.Close()
	return nil
}

func (c *ClientEphemeralfiles) DownloadPartE2EEndpoint(transactionID string, part int) string {
	return fmt.Sprintf("%s/%s/multipart/%s/%d", c.endpoint, apiVersion, transactionID, part)
}

func (c *ClientEphemeralfiles) DownloadPartE2E(outputFilePath string, transactionID string, aesKey []byte, part int) (int, error) {
	req, err := http.NewRequest(http.MethodGet, c.DownloadPartE2EEndpoint(transactionID, part), nil)
	if err != nil {
		return 0, fmt.Errorf("error creating request: %w", err)
	}
	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading response: %w", err)
	}
	// Decrypt response
	decryptedChunk, err := DecryptAES(aesKey, body)
	if err != nil {
		return 0, fmt.Errorf("error decrypting chunk: %w", err)
	}

	// Write decrypted chunk to file
	file, err := os.OpenFile(outputFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return 0, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(decryptedChunk)
	if err != nil {
		return 0, fmt.Errorf("error writing chunk to file: %w", err)
	}

	stat, err := file.Stat()
	if err != nil {
		return 0, fmt.Errorf("error getting file info: %w", err)
	}
	return int(stat.Size()), nil
}

func (c *ClientEphemeralfiles) UpdateAESKeyForDownloadTransactionEndpoint(transactionID string) string {
	return fmt.Sprintf("%s/%s/download-transaction/%s/aeskey", c.endpoint, apiVersion, transactionID)
}

func (c *ClientEphemeralfiles) SendAESKeyForDownloadTransaction(transactionID, encryptedAESKey string) error {
	payload := dto.RequestAESKey{
		AESKey: encryptedAESKey,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshalling payload: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, c.UpdateAESKeyForDownloadTransactionEndpoint(transactionID), bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func DecryptAES(key []byte, ciphertext []byte) ([]byte, error) {
	// Check if ciphertext is too short
	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	// Create new cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	// Extract IV from first BlockSize bytes of ciphertext
	iv := ciphertext[:aes.BlockSize]
	// Create CTR stream
	stream := cipher.NewCTR(block, iv)
	// Create buffer for plaintext
	plaintext := make([]byte, len(ciphertext)-aes.BlockSize)
	// Decrypt ciphertext
	stream.XORKeyStream(plaintext, ciphertext[aes.BlockSize:])
	return plaintext, nil
}
