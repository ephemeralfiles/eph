package ephcli

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

func (c *ClientEphemeralfiles) SendAESKeyEndpoint(fileID string) string {
	return fmt.Sprintf("%s/%s/files/%s/key", c.endpoint, apiVersion, fileID)
}

func (c *ClientEphemeralfiles) GetPublicKeyEndpoint() string {
	return fmt.Sprintf("%s/%s/files", c.endpoint, apiVersion)
}

func (c *ClientEphemeralfiles) UploadE2EEndpoint(uuid string) string {
	return fmt.Sprintf("%s/%s/multipart-upload/%s", c.endpoint, apiVersion, uuid)
}

const (
	chunkSize = 1024 * 1024 // 1MB chunks
)

func (c *ClientEphemeralfiles) GetPublicKey() (string, string, error) {
	req, err := http.NewRequest("HEAD", c.GetPublicKeyEndpoint(), nil)
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
	fmt.Println("Uploading file in chunks", targetURL)
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

	// Create HTTP client
	client := &http.Client{}

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
		req, err := http.NewRequest("POST", targetURL, body)
		if err != nil {
			return fmt.Errorf("error creating request: %w", err)
		}

		// Set headers
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))

		// Send request
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("error sending request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		fmt.Printf("Uploaded chunk %d-%d of %d bytes\n", start, end, fileSize)
	}

	return nil
}

func (c *ClientEphemeralfiles) UploadE2E(fileToUpload string) error {
	// if c.noProgressBar {
	// 	return c.UploadWithoutProgressBar(fileToUpload)
	// }
	// return uploadFileInChunks(fileToUpload, c.UploadE2EEndpoint(uuid))
	return nil
}

func (c *ClientEphemeralfiles) SendAESKey(fileID, encryptedAESKey string) error {

	type DTORequestAESKey struct {
		AESKey string `json:"aeskey"`
	}
	payload := DTORequestAESKey{
		AESKey: encryptedAESKey,
	}
	// Send request
	// marshal payload
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshalling payload: %w", err)
	}
	req, err := http.NewRequest("POST", c.SendAESKeyEndpoint(fileID), bytes.NewBuffer(body))
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
		return nil, err
	}

	// Generate random IV
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
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
