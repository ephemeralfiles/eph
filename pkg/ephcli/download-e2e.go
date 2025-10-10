package ephcli

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/ephemeralfiles/eph/pkg/dto"
)

const (
	// FilePermission is the permission for downloaded files.
	FilePermission = 0600
)

// GetFileInformationEndpoint returns the API endpoint URL for retrieving file information.
func (c *ClientEphemeralfiles) GetFileInformationEndpoint(fileID string) string {
	return fmt.Sprintf("%s/%s/files/info/%s", c.endpoint, apiVersion, fileID)
}

// GetNewDownloadTransactionEndpoint returns the API endpoint URL for creating a new download transaction.
func (c *ClientEphemeralfiles) GetNewDownloadTransactionEndpoint(fileID string) string {
	return fmt.Sprintf("%s/%s/download/encrypted/%s/init", c.endpoint, apiVersion, fileID)
}

// CreateNewDownloadTransaction creates a new E2E download transaction and returns the transaction ID and public key.
func (c *ClientEphemeralfiles) CreateNewDownloadTransaction(
	fileID string,
) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultAPIRequestTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.GetNewDownloadTransactionEndpoint(fileID), nil)
	if err != nil {
		return "", "", fmt.Errorf("error creating request: %w", err)
	}
	// Set headers
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("error sending request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("%w: %d", ErrUnexpectedStatusCode, resp.StatusCode)
	}

	// Get Header X-File-Id from Header
	transactionID := resp.Header.Get("X-Transaction-Id")
	if transactionID == "" {
		return "", "", fmt.Errorf("error reading response: %w", err)
	}
	// Get Header X-File-Public-Key from Header
	publicKey := resp.Header.Get("X-File-Public-Key")
	if publicKey == "" {
		return "", "", fmt.Errorf("error reading response: %w", err)
	}
	return transactionID, publicKey, nil
}

// DownloadE2E downloads and decrypts a file using end-to-end encryption.
func (c *ClientEphemeralfiles) DownloadE2E(fileID string, outputPath string) error {
	// Get file information
	fileInfo, err := c.getFileInformation(fileID)
	if err != nil {
		return err
	}
	c.logFileInfo(fileInfo)

	// Determine output file path: use provided path or fallback to server filename
	outputFilePath := outputPath
	if outputFilePath == "" {
		outputFilePath = fileInfo.Filename
	}

	// Setup download transaction and encryption
	transactionID, keyBundle, err := c.setupDownloadTransaction(fileID)
	if err != nil {
		return err
	}

	// Download all parts
	return c.downloadAllParts(fileInfo, transactionID, keyBundle.AESKey, outputFilePath)
}

// DownloadPartE2EEndpoint returns the API endpoint URL for downloading a specific part of an E2E encrypted file.
func (c *ClientEphemeralfiles) DownloadPartE2EEndpoint(transactionID string, part int) string {
	return fmt.Sprintf("%s/%s/download/encrypted/%s/chunks/%d", c.endpoint, apiVersion, transactionID, part)
}

// DownloadPartE2E downloads and decrypts a specific part of an E2E encrypted file.
// Deprecated: Use DownloadPartE2EToFile instead.
func (c *ClientEphemeralfiles) DownloadPartE2E(
	outputFilePath string, transactionID string, aesKey []byte, part int,
) (int, error) {
	file, err := os.OpenFile(outputFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, FilePermission)
	if err != nil {
		return 0, fmt.Errorf("error opening file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	return c.DownloadPartE2EToFile(file, transactionID, aesKey, part)
}

// DownloadPartE2EToFile downloads and decrypts a specific part of an E2E encrypted file to an open file handle.
func (c *ClientEphemeralfiles) DownloadPartE2EToFile(
	file *os.File, transactionID string, aesKey []byte, part int,
) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultAPIRequestTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.DownloadPartE2EEndpoint(transactionID, part), nil)
	if err != nil {
		return 0, fmt.Errorf("error creating request: %w", err)
	}
	// Set headers
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error sending request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("%w: %d", ErrUnexpectedStatusCode, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading response: %w", err)
	}

	encryptedSize := len(body)
	c.log.Debug("Received encrypted chunk",
		slog.Int("part", part),
		slog.Int("encryptedSize", encryptedSize))

	// Decrypt response
	decryptedChunk, err := DecryptAES(aesKey, body)
	if err != nil {
		return 0, fmt.Errorf("error decrypting chunk: %w", err)
	}

	decryptedSize := len(decryptedChunk)
	c.log.Debug("Decrypted chunk",
		slog.Int("part", part),
		slog.Int("decryptedSize", decryptedSize),
		slog.Int("ivOverhead", encryptedSize-decryptedSize))

	// Write the chunk directly to the already-open file
	n, err := file.Write(decryptedChunk)
	if err != nil {
		return 0, fmt.Errorf("error writing chunk to file: %w", err)
	}

	c.log.Debug("Wrote chunk to file",
		slog.Int("part", part),
		slog.Int("bytesWritten", n),
		slog.Int("expectedToWrite", decryptedSize))

	// Update progress bar
	if c.bar != nil {
		_, _ = c.bar.Write(decryptedChunk)
	}

	return n, nil
}

// UpdateAESKeyForDownloadTransactionEndpoint returns the API endpoint URL for updating
// the AES key in a download transaction.
func (c *ClientEphemeralfiles) UpdateAESKeyForDownloadTransactionEndpoint(transactionID string) string {
	return fmt.Sprintf("%s/%s/download/encrypted/%s/key", c.endpoint, apiVersion, transactionID)
}


// DecryptAES decrypts ciphertext using AES decryption with the provided key.
func DecryptAES(key []byte, ciphertext []byte) ([]byte, error) {
	// Check if ciphertext is too short
	if len(ciphertext) < aes.BlockSize {
		return nil, ErrCiphertextTooShort
	}
	// Create new cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreatingCipherBlock, err)
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

// getFileInformation retrieves file information from the server.
func (c *ClientEphemeralfiles) getFileInformation(fileID string) (*dto.InfoFile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultAPIRequestTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.GetFileInformationEndpoint(fileID), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", ErrUnexpectedStatusCode, resp.StatusCode)
	}

	var fileInfo dto.InfoFile
	err = json.NewDecoder(resp.Body).Decode(&fileInfo)
	if err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}
	return &fileInfo, nil
}

// logFileInfo logs file information for debugging.
func (c *ClientEphemeralfiles) logFileInfo(fileInfo *dto.InfoFile) {
	c.log.Debug("DownloadE2E",
		slog.String("Filename", fileInfo.Filename),
		slog.Int("NbParts", fileInfo.NbParts),
		slog.Int64("Size", fileInfo.Size))
}

// setupDownloadTransaction creates download transaction and encryption keys.
func (c *ClientEphemeralfiles) setupDownloadTransaction(fileID string) (string, *E2EKeyBundle, error) {
	transactionID, pubkey, err := c.CreateNewDownloadTransaction(fileID)
	if err != nil {
		return "", nil, fmt.Errorf("error creating new download transaction: %w", err)
	}
	c.log.Debug("DownloadE2E", slog.String("TransactionID", transactionID), slog.String("PublicKey", pubkey))

	keyBundle, err := GenerateAndEncryptAESKey(pubkey)
	if err != nil {
		return "", nil, fmt.Errorf("error generating and encrypting AES key: %w", err)
	}

	c.log.Debug("DownloadE2E",
		slog.String("AESKey", string(keyBundle.AESKey)),
		slog.String("HexString", keyBundle.HexString))
	c.log.Debug("DownloadE2E", slog.String("EncryptedAESKey", keyBundle.EncryptedAESKey))

	err = c.SendAESKeyToEndpoint(c.UpdateAESKeyForDownloadTransactionEndpoint(transactionID), keyBundle.EncryptedAESKey)
	if err != nil {
		return "", nil, fmt.Errorf("error sending AES key: %w", err)
	}

	return transactionID, keyBundle, nil
}

// downloadAllParts downloads all file parts with progress tracking.
func (c *ClientEphemeralfiles) downloadAllParts(fileInfo *dto.InfoFile, transactionID string, aesKey []byte, outputFilePath string) error {
	c.InitProgressBar("downloading file...", fileInfo.Size)
	defer c.CloseProgressBar()

	// Open file once for all chunks
	// #nosec G304 -- outputFilePath is controlled by user for file download
	file, err := os.OpenFile(outputFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, FilePermission)
	if err != nil {
		return fmt.Errorf("error opening output file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	c.log.Info("Starting download",
		slog.Int("totalParts", fileInfo.NbParts),
		slog.Int64("expectedSize", fileInfo.Size))

	totalBytesWritten := int64(0)
	for i := range fileInfo.NbParts {
		c.log.Debug("DownloadE2E", slog.Int("Part", i))
		chunkSize, err := c.DownloadPartE2EToFile(file, transactionID, aesKey, i)
		if err != nil {
			return fmt.Errorf("error downloading part %d: %w", i, err)
		}
		totalBytesWritten += int64(chunkSize)
		c.log.Info("Downloaded chunk",
			slog.Int("part", i),
			slog.Int("chunkSize", chunkSize),
			slog.Int64("totalWritten", totalBytesWritten),
			slog.Int64("expected", fileInfo.Size))
	}

	c.log.Info("Download complete",
		slog.Int64("totalBytesWritten", totalBytesWritten),
		slog.Int64("expectedSize", fileInfo.Size),
		slog.Int64("difference", fileInfo.Size-totalBytesWritten))

	return nil
}
