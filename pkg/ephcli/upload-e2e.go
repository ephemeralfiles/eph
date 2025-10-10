package ephcli

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
)

const (
	chunkSize = 128 * 1024 * 1024 // 128MB chunks
)

// SendAESKeyEndpoint returns the API endpoint URL for sending an AES key for a file.
func (c *ClientEphemeralfiles) SendAESKeyEndpoint(uploadID string) string {
	return fmt.Sprintf("%s/%s/upload/encrypted/%s/key", c.endpoint, apiVersion, uploadID)
}

// GetPublicKeyEndpoint returns the API endpoint URL for retrieving the server's public key.
func (c *ClientEphemeralfiles) GetPublicKeyEndpoint() string {
	return fmt.Sprintf("%s/%s/upload/encrypted/init", c.endpoint, apiVersion)
}

// UploadE2EEndpoint returns the API endpoint URL for E2E encrypted file uploads.
func (c *ClientEphemeralfiles) UploadE2EEndpoint(uploadID string) string {
	return fmt.Sprintf("%s/%s/upload/encrypted/%s/chunks", c.endpoint, apiVersion, uploadID)
}

// GetPublicKey retrieves the server's public key and creates a new upload transaction.
func (c *ClientEphemeralfiles) GetPublicKey() (string, string, string, error) {
	return c.GetPublicKeyWithHeaders("", nil)
}

// GetPublicKeyWithHeaders retrieves the server's public key with optional organization context.
func (c *ClientEphemeralfiles) GetPublicKeyWithHeaders(orgID string, tags []string) (string, string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultAPIRequestTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.GetPublicKeyEndpoint(), nil)
	if err != nil {
		return "", "", "", fmt.Errorf("error creating request: %w", err)
	}
	// Set headers
	req.Header.Set("Authorization", "Bearer "+c.token)

	// Add organization headers if provided
	if orgID != "" {
		req.Header.Set("X-Organization-Id", orgID)
	}
	if len(tags) > 0 {
		req.Header.Set("X-File-Tags", joinTags(tags))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", "", fmt.Errorf("error sending request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.log.Debug("Warning: failed to close response body", slog.String("error", closeErr.Error()))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", "", "", fmt.Errorf("%w: %d", ErrUnexpectedStatusCode, resp.StatusCode)
	}

	// Get Header X-File-Id from Header
	fileID := resp.Header.Get("X-File-Id")
	if fileID == "" {
		return "", "", "", fmt.Errorf("error reading response: %w", err)
	}
	// Get Header X-File-Public-Key from Header
	publicKey := resp.Header.Get("X-File-Public-Key")
	if publicKey == "" {
		return "", "", "", fmt.Errorf("error reading response: %w", err)
	}
	transactionID := resp.Header.Get("X-Upload-Id")
	if transactionID == "" {
		return "", "", "", fmt.Errorf("error reading response: %w", err)
	}

	c.log.Debug("GetPublicKey", slog.String("X-File-Public-Key", publicKey))
	c.log.Debug("GetPublicKey", slog.String("X-File-Id", fileID))
	c.log.Debug("GetPublicKey", slog.String("X-Upload-Id", transactionID))
	return transactionID, fileID, publicKey, nil
}

// UploadFileInChunks uploads a file in encrypted chunks for E2E encryption.
func (c *ClientEphemeralfiles) UploadFileInChunks(aeskey []byte, filePath, targetURL string) error {
	c.log.Debug("UploadFileInChunks", slog.String("aeskey", string(aeskey)))
	c.log.Debug("UploadFileInChunks", slog.String("filePath", filePath))
	c.log.Debug("UploadFileInChunks", slog.String("targetURL", targetURL))
	
	file, fileSize, err := c.openFileForUpload(filePath)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	// Create progress bar
	c.InitProgressBar("uploading file...", fileSize)
	defer c.CloseProgressBar()

	// Upload file in chunks
	for start := int64(0); start < fileSize; start += chunkSize {
		end := c.calculateChunkEnd(start, fileSize)
		if err := c.uploadSingleChunk(file, aeskey, targetURL, start, end, fileSize); err != nil {
			return err
		}
		_ = c.bar.Add(chunkSize)
	}
	return nil
}

// UploadE2E uploads a file using end-to-end encryption.
func (c *ClientEphemeralfiles) UploadE2E(fileToUpload string) error {
	transactionID, fileID, pubkey, err := c.GetPublicKey()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting public key: %s\n", err.Error())
		os.Exit(1)
	}
	c.log.Debug("UploadE2E", slog.String("fileID", fileID))
	c.log.Debug("UploadE2E", slog.String("pubkey", pubkey))

	// Generate and encrypt AES key using shared utility
	keyBundle, err := GenerateAndEncryptAESKey(pubkey)
	if err != nil {
		return fmt.Errorf("error generating and encrypting AES key: %w", err)
	}

	c.log.Debug("UploadE2E", slog.String("aesKey", string(keyBundle.AESKey)))
	c.log.Debug("UploadE2E", slog.String("hexString", keyBundle.HexString))
	c.log.Debug("UploadE2E", slog.String("encryptedAESKey", keyBundle.EncryptedAESKey))
	c.log.Debug("UploadE2E", slog.String("fileToUpload", fileToUpload))

	// Send the encrypted AES key to the server using shared utility
	err = c.SendAESKeyToEndpoint(c.SendAESKeyEndpoint(transactionID), keyBundle.EncryptedAESKey)
	if err != nil {
		return fmt.Errorf("error sending AES key: %w", err)
	}

	// Upload the file
	err = c.UploadFileInChunks(keyBundle.AESKey, fileToUpload, c.UploadE2EEndpoint(transactionID))
	if err != nil {
		return fmt.Errorf("error uploading file: %w", err)
	}
	return nil
}

// UploadOrganizationFileE2E uploads a file to an organization using end-to-end encryption.
func (c *ClientEphemeralfiles) UploadOrganizationFileE2E(orgID string, fileToUpload string, tags []string) (string, error) {
	transactionID, fileID, pubkey, err := c.GetPublicKeyWithHeaders(orgID, tags)
	if err != nil {
		return "", fmt.Errorf("error getting public key: %w", err)
	}
	c.log.Debug("UploadOrganizationFileE2E", slog.String("fileID", fileID), slog.String("orgID", orgID))
	c.log.Debug("UploadOrganizationFileE2E", slog.String("pubkey", pubkey))

	// Generate and encrypt AES key using shared utility
	keyBundle, err := GenerateAndEncryptAESKey(pubkey)
	if err != nil {
		return "", fmt.Errorf("error generating and encrypting AES key: %w", err)
	}

	c.log.Debug("UploadOrganizationFileE2E", slog.String("aesKey", string(keyBundle.AESKey)))
	c.log.Debug("UploadOrganizationFileE2E", slog.String("hexString", keyBundle.HexString))
	c.log.Debug("UploadOrganizationFileE2E", slog.String("encryptedAESKey", keyBundle.EncryptedAESKey))
	c.log.Debug("UploadOrganizationFileE2E", slog.String("fileToUpload", fileToUpload))

	// Send the encrypted AES key to the server using shared utility
	err = c.SendAESKeyToEndpoint(c.SendAESKeyEndpoint(transactionID), keyBundle.EncryptedAESKey)
	if err != nil {
		return "", fmt.Errorf("error sending AES key: %w", err)
	}

	// Upload the file
	err = c.UploadFileInChunks(keyBundle.AESKey, fileToUpload, c.UploadE2EEndpoint(transactionID))
	if err != nil {
		return "", fmt.Errorf("error uploading file: %w", err)
	}

	return fileID, nil
}


// EncryptAES encrypts plaintext using AES encryption with the provided key.
func EncryptAES(key []byte, plaintext []byte) ([]byte, error) {
	// Create new cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("error creating new cipher block: %w", err)
	}
	// Generate random IV
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("error generating random IV: %w", err)
	}
	// Create CTR stream
	stream := cipher.NewCTR(block, iv)
	// Create buffer for ciphertext that includes space for IV
	ciphertext := make([]byte, len(iv)+len(plaintext))
	// Copy IV to start of ciphertext
	copy(ciphertext[:aes.BlockSize], iv)
	// Encrypt plaintext
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)
	return ciphertext, nil
}

// openFileForUpload opens a file and returns file handle and size.
func (c *ClientEphemeralfiles) openFileForUpload(filePath string) (*os.File, int64, error) {
	// #nosec G304 -- filePath is provided by user for file upload
	file, err := os.Open(filePath)
	if err != nil {
		return nil, 0, fmt.Errorf("error opening file: %w", err)
	}

	fileInfo, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, 0, fmt.Errorf("error getting file info: %w", err)
	}

	return file, fileInfo.Size(), nil
}

// calculateChunkEnd calculates the end position for a chunk.
func (c *ClientEphemeralfiles) calculateChunkEnd(start, fileSize int64) int64 {
	end := start + chunkSize - 1
	if end >= fileSize {
		return fileSize - 1
	}
	return end
}

// uploadSingleChunk uploads a single encrypted chunk.
func (c *ClientEphemeralfiles) uploadSingleChunk(
	file *os.File, aeskey []byte, targetURL string, start, end, fileSize int64,
) error {
	// Read chunk
	chunkSize := end - start + 1
	chunk := make([]byte, chunkSize)
	bytesRead, err := file.ReadAt(chunk, start)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("error reading chunk: %w", err)
	}

	c.log.Debug("Read chunk",
		slog.Int64("start", start),
		slog.Int64("end", end),
		slog.Int("expectedSize", int(chunkSize)),
		slog.Int("actualBytesRead", bytesRead))

	// Use only the bytes actually read
	actualChunk := chunk[:bytesRead]

	// Encrypt chunk
	encryptedChunk, err := EncryptAES(aeskey, actualChunk)
	if err != nil {
		return fmt.Errorf("error encrypting chunk: %w", err)
	}

	c.log.Debug("Encrypted chunk",
		slog.Int("plaintextSize", bytesRead),
		slog.Int("encryptedSize", len(encryptedChunk)))

	// Create multipart form
	body, contentType, err := c.createChunkForm(encryptedChunk, file)
	if err != nil {
		return err
	}

	// Send request
	return c.sendChunkRequest(targetURL, body, contentType, start, end, fileSize)
}

// createChunkForm creates multipart form for chunk upload.
func (c *ClientEphemeralfiles) createChunkForm(encryptedChunk []byte, file *os.File) (*bytes.Buffer, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, "", fmt.Errorf("error getting file info: %w", err)
	}

	part, err := writer.CreateFormFile("uploadfile", fileInfo.Name())
	if err != nil {
		return nil, "", fmt.Errorf("error creating form file: %w", err)
	}

	_, err = part.Write(encryptedChunk)
	if err != nil {
		return nil, "", fmt.Errorf("error writing chunk to form: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return nil, "", fmt.Errorf("error closing writer: %w", err)
	}

	return body, writer.FormDataContentType(), nil
}

// sendChunkRequest sends HTTP request for chunk upload.
func (c *ClientEphemeralfiles) sendChunkRequest(
	targetURL string, body *bytes.Buffer, contentType string, start, end, fileSize int64,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), ChunkUploadTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, body)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.log.Debug("Warning: failed to close response body", slog.String("error", closeErr.Error()))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %d", ErrUnexpectedStatusCode, resp.StatusCode)
	}

	c.log.Debug("UploadFileInChunks", slog.Int64("start", start), slog.Int64("end", end), slog.Int64("fileSize", fileSize))
	return nil
}

// joinTags converts a slice of tags to a comma-separated string.
func joinTags(tags []string) string {
	result := ""
	for i, tag := range tags {
		if i > 0 {
			result += ","
		}
		result += tag
	}
	return result
}
