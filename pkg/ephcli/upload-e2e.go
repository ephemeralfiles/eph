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
	"mime/multipart"
	"net/http"
	"os"

	"github.com/ephemeralfiles/eph/pkg/dto"
	"github.com/schollz/progressbar/v3"
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

const (
	chunkSize = 128 * 1024 * 1024 // 128MB chunks
)

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

	// Create progress bar
	bar := progressbar.NewOptions64(fileSize, progressbar.OptionClearOnFinish(),
		progressbar.OptionShowBytes(true), progressbar.OptionSetWidth(DefaultBarWidth),
		progressbar.OptionSetDescription("uploadding file..."),
		progressbar.OptionSetVisibility(!c.noProgressBar),
	)

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
		req, err := http.NewRequest(http.MethodPost, targetURL, body)
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

		_ = bar.Add(chunkSize)
		// fmt.Printf("Uploaded chunk %d-%d of %d bytes\n", start, end, fileSize)
	}
	bar.Clear()
	bar.Close()
	return nil
}

func (c *ClientEphemeralfiles) UploadE2E(fileToUpload string) error {
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
