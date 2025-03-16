package ephcli

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/ephemeralfiles/eph/pkg/dto"
)

const (
	chunkSize = 128 * 1024 * 1024 // 128MB chunks
)

func (c *ClientEphemeralfiles) SendAESKeyEndpoint(fileID string) string {
	return fmt.Sprintf("%s/%s/files/%s/upload-key", c.endpoint, apiVersion, fileID)
}

func (c *ClientEphemeralfiles) GetPublicKeyEndpoint() string {
	return fmt.Sprintf("%s/%s/files", c.endpoint, apiVersion)
}

func (c *ClientEphemeralfiles) UploadE2EEndpoint(uuid string) string {
	return fmt.Sprintf("%s/%s/multipart/%s", c.endpoint, apiVersion, uuid)
}

func (c *ClientEphemeralfiles) GetPublicKey() (string, string, error) {
	req, err := http.NewRequest(http.MethodHead, c.GetPublicKeyEndpoint(), nil)
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
	fileId := resp.Header.Get("X-File-Id")
	if fileId == "" {
		return "", "", fmt.Errorf("error reading response: %w", err)
	}
	// Get Header X-File-Public-Key from Header
	publicKey := resp.Header.Get("X-File-Public-Key")
	if publicKey == "" {
		return "", "", fmt.Errorf("error reading response: %w", err)
	}
	return fileId, publicKey, nil
}

func (c *ClientEphemeralfiles) UploadFileInChunks(aeskey []byte, filePath, targetURL string) error {
	c.log.Debug("UploadFileInChunks", slog.String("aeskey", string(aeskey)))
	c.log.Debug("UploadFileInChunks", slog.String("filePath", filePath))
	c.log.Debug("UploadFileInChunks", slog.String("targetURL", targetURL))
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("error getting file info: %w", err)
	}
	fileSize := fileInfo.Size()

	// Create progress bar
	c.InitProgressBar("uploading file...", fileSize)
	defer c.CloseProgressBar()

	// Upload file in chunks
	for start := int64(0); start < fileSize; start += chunkSize {
		// Calculate end of chunk
		end := start + chunkSize - 1
		if end >= fileSize {
			end = fileSize - 1
		}

		// Create buffer for chunk
		chunk := make([]byte, end-start+1)
		_, err = file.ReadAt(chunk, start)
		if err != nil && err != io.EOF {
			return fmt.Errorf("error reading chunk: %w", err)
		}

		// Encrypt chunk
		encryptedChunk, err := EncryptAES(aeskey, chunk)
		if err != nil {
			return fmt.Errorf("error encrypting chunk: %w", err)
		}

		// Create multipart form
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		part, err := writer.CreateFormFile("uploadfile", fileInfo.Name())
		if err != nil {
			return fmt.Errorf("error creating form file: %w", err)
		}

		_, err = part.Write(encryptedChunk)
		if err != nil {
			return fmt.Errorf("error writing chunk to form: %w", err)
		}

		err = writer.Close()
		if err != nil {
			return fmt.Errorf("error closing writer: %w", err)
		}

		// Create request
		req, err := http.NewRequest(http.MethodPost, targetURL, body)
		if err != nil {
			return fmt.Errorf("error creating request: %w", err)
		}

		// Set headers
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))

		// Send request
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("error sending request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		_ = c.bar.Add(chunkSize)
		c.log.Debug("UploadFileInChunks", slog.Int64("start", start))
		c.log.Debug("UploadFileInChunks", slog.Int64("end", end))
		c.log.Debug("UploadFileInChunks", slog.Int64("fileSize", fileSize))
		c.log.Debug("UploadFileInChunks", slog.Int("chunkSize", len(encryptedChunk)))
	}
	return nil
}

func (c *ClientEphemeralfiles) UploadE2E(fileToUpload string) error {
	fileID, pubkey, err := c.GetPublicKey()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting public key: %s\n", err)
		os.Exit(1)
	}
	c.log.Debug("UploadFileInChunks", slog.String("fileID", fileID))
	c.log.Debug("UploadFileInChunks", slog.String("pubkey", pubkey))

	aesKey, err := GenAESKey32bits()
	if err != nil {
		return fmt.Errorf("error generating AES key: %w", err)
	}
	c.log.Debug("UploadFileInChunks", slog.String("aesKey", string(aesKey)))
	c.log.Debug("UploadFileInChunks", slog.String("fileToUpload", fileToUpload))
	hexString := hex.EncodeToString(aesKey)
	c.log.Debug("UploadFileInChunks", slog.String("hexString", hexString))

	// encrypt with public key
	encryptedAESKey, err := EncryptAESKey(pubkey, hexString)
	if err != nil {
		return fmt.Errorf("error encrypting AES key: %w", err)
	}
	c.log.Debug("UploadFileInChunks", slog.String("encryptedAESKey", encryptedAESKey))
	// Send the encrypted AES key to the server
	err = c.SendAESKey(fileID, encryptedAESKey)
	if err != nil {
		return fmt.Errorf("error sending AES key: %w", err)
	}
	// Upload the file
	err = c.UploadFileInChunks(aesKey, fileToUpload, c.UploadE2EEndpoint(fileID))
	if err != nil {
		return fmt.Errorf("error uploading file: %w", err)
	}
	return nil
}

func (c *ClientEphemeralfiles) SendAESKey(fileID, encryptedAESKey string) error {
	payload := dto.RequestAESKey{
		AESKey: encryptedAESKey,
	}
	// Send request
	// marshal payload
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshalling payload: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, c.SendAESKeyEndpoint(fileID), bytes.NewBuffer(body))
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
