package ephcli

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"strings"

	"github.com/ephemeralfiles/eph/pkg/dto"
)

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

func LoadRSAPublicKey(pemString string) (*rsa.PublicKey, error) {
	formattedPEM := formatPEM(pemString)
	
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

// GenAESKey32bits generates a 32 bits AES key
func GenAESKey32bits() ([]byte, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("error generating AES key: %w", err)
	}
	return key, nil
}

// E2EKeyBundle represents an AES key with its encrypted version
type E2EKeyBundle struct {
	AESKey          []byte
	HexString       string
	EncryptedAESKey string
}

// GenerateAndEncryptAESKey creates an AES key and encrypts it with the given public key
func GenerateAndEncryptAESKey(publicKey string) (*E2EKeyBundle, error) {
	aesKey, err := GenAESKey32bits()
	if err != nil {
		return nil, fmt.Errorf("error generating AES key: %w", err)
	}

	hexString := hex.EncodeToString(aesKey)
	
	encryptedAESKey, err := EncryptAESKey(publicKey, hexString)
	if err != nil {
		return nil, fmt.Errorf("error encrypting AES key: %w", err)
	}

	return &E2EKeyBundle{
		AESKey:          aesKey,
		HexString:       hexString,
		EncryptedAESKey: encryptedAESKey,
	}, nil
}

// SendAESKeyToEndpoint sends an encrypted AES key to the specified endpoint
func (c *ClientEphemeralfiles) SendAESKeyToEndpoint(endpoint, encryptedAESKey string) error {
	payload := dto.RequestAESKey{
		AESKey: encryptedAESKey,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshalling payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(body))
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
