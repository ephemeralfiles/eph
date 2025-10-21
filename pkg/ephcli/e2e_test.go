package ephcli_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ephemeralfiles/eph/pkg/ephcli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateTestRSAKeyPair generates a test RSA key pair for testing.
func generateTestRSAKeyPair(t *testing.T) (*rsa.PrivateKey, string) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return privateKey, string(publicKeyPEM)
}

func TestGenAESKey32bits(t *testing.T) {
	t.Parallel()

	t.Run("generates 32-byte key", func(t *testing.T) {
		t.Parallel()

		key, err := ephcli.GenAESKey32bits()
		require.NoError(t, err)
		assert.Len(t, key, 32)
	})

	t.Run("generates unique keys", func(t *testing.T) {
		t.Parallel()

		key1, err := ephcli.GenAESKey32bits()
		require.NoError(t, err)

		key2, err := ephcli.GenAESKey32bits()
		require.NoError(t, err)

		// Keys should be different (extremely unlikely to be equal)
		assert.NotEqual(t, key1, key2)
	})

	t.Run("generates non-zero keys", func(t *testing.T) {
		t.Parallel()

		key, err := ephcli.GenAESKey32bits()
		require.NoError(t, err)

		// Check that not all bytes are zero
		allZero := true
		for _, b := range key {
			if b != 0 {
				allZero = false
				break
			}
		}
		assert.False(t, allZero, "Generated key should not be all zeros")
	})
}

func TestLoadRSAPublicKey(t *testing.T) {
	t.Parallel()

	t.Run("loads valid PEM public key", func(t *testing.T) {
		t.Parallel()

		_, publicKeyPEM := generateTestRSAKeyPair(t)

		pubKey, err := ephcli.LoadRSAPublicKey(publicKeyPEM)
		require.NoError(t, err)
		require.NotNil(t, pubKey)
		assert.IsType(t, &rsa.PublicKey{}, pubKey)
	})

	t.Run("loads PEM without line breaks", func(t *testing.T) {
		t.Parallel()

		_, publicKeyPEM := generateTestRSAKeyPair(t)

		// Remove line breaks to simulate single-line PEM
		singleLinePEM := string(publicKeyPEM)

		pubKey, err := ephcli.LoadRSAPublicKey(singleLinePEM)
		require.NoError(t, err)
		require.NotNil(t, pubKey)
	})

	t.Run("error on invalid PEM", func(t *testing.T) {
		t.Parallel()

		invalidPEM := "this is not a valid PEM"

		pubKey, err := ephcli.LoadRSAPublicKey(invalidPEM)
		require.Error(t, err)
		assert.Nil(t, pubKey)
	})

	t.Run("error on malformed PEM block", func(t *testing.T) {
		t.Parallel()

		malformedPEM := `-----BEGIN PUBLIC KEY-----
invalid base64 content here!@#$
-----END PUBLIC KEY-----`

		pubKey, err := ephcli.LoadRSAPublicKey(malformedPEM)
		require.Error(t, err)
		assert.Nil(t, pubKey)
	})

	t.Run("error on wrong key type", func(t *testing.T) {
		t.Parallel()

		// Create a non-RSA key (private key instead)
		wrongTypePEM := `-----BEGIN PRIVATE KEY-----
MIIBVAIBADANBgkqhkiG9w0BAQEFAASCAT4wggE6AgEAAkEA
-----END PRIVATE KEY-----`

		pubKey, err := ephcli.LoadRSAPublicKey(wrongTypePEM)
		require.Error(t, err)
		assert.Nil(t, pubKey)
	})
}

func TestEncryptAESKey(t *testing.T) {
	t.Parallel()

	t.Run("successful encryption", func(t *testing.T) {
		t.Parallel()

		_, publicKeyPEM := generateTestRSAKeyPair(t)

		// Generate a test AES key
		aesKey, err := ephcli.GenAESKey32bits()
		require.NoError(t, err)
		aesKeyHex := hex.EncodeToString(aesKey)

		encryptedKey, err := ephcli.EncryptAESKey(publicKeyPEM, aesKeyHex)
		require.NoError(t, err)
		assert.NotEmpty(t, encryptedKey)
	})

	t.Run("encrypted keys are different each time", func(t *testing.T) {
		t.Parallel()

		_, publicKeyPEM := generateTestRSAKeyPair(t)

		aesKey, err := ephcli.GenAESKey32bits()
		require.NoError(t, err)
		aesKeyHex := hex.EncodeToString(aesKey)

		encrypted1, err := ephcli.EncryptAESKey(publicKeyPEM, aesKeyHex)
		require.NoError(t, err)

		encrypted2, err := ephcli.EncryptAESKey(publicKeyPEM, aesKeyHex)
		require.NoError(t, err)

		// Due to OAEP padding, encryptions should be different
		assert.NotEqual(t, encrypted1, encrypted2)
	})

	t.Run("error on invalid public key", func(t *testing.T) {
		t.Parallel()

		invalidPEM := "invalid public key"
		aesKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

		encryptedKey, err := ephcli.EncryptAESKey(invalidPEM, aesKey)
		require.Error(t, err)
		assert.Empty(t, encryptedKey)
	})

	t.Run("handles PEM without line breaks", func(t *testing.T) {
		t.Parallel()

		_, publicKeyPEM := generateTestRSAKeyPair(t)

		// Remove all line breaks
		singleLinePEM := string(publicKeyPEM)

		aesKey, err := ephcli.GenAESKey32bits()
		require.NoError(t, err)
		aesKeyHex := hex.EncodeToString(aesKey)

		encryptedKey, err := ephcli.EncryptAESKey(singleLinePEM, aesKeyHex)
		require.NoError(t, err)
		assert.NotEmpty(t, encryptedKey)
	})
}

func TestGenerateAndEncryptAESKey(t *testing.T) {
	t.Parallel()

	t.Run("successful generation and encryption", func(t *testing.T) {
		t.Parallel()

		_, publicKeyPEM := generateTestRSAKeyPair(t)

		bundle, err := ephcli.GenerateAndEncryptAESKey(publicKeyPEM)
		require.NoError(t, err)
		require.NotNil(t, bundle)

		// Verify bundle components
		assert.Len(t, bundle.AESKey, 32, "AES key should be 32 bytes")
		assert.NotEmpty(t, bundle.HexString, "Hex string should not be empty")
		assert.NotEmpty(t, bundle.EncryptedAESKey, "Encrypted AES key should not be empty")

		// Verify hex string is valid hex representation
		decodedKey, err := hex.DecodeString(bundle.HexString)
		require.NoError(t, err)
		assert.Equal(t, bundle.AESKey, decodedKey, "Hex string should decode to AES key")
	})

	t.Run("generates unique bundles", func(t *testing.T) {
		t.Parallel()

		_, publicKeyPEM := generateTestRSAKeyPair(t)

		bundle1, err := ephcli.GenerateAndEncryptAESKey(publicKeyPEM)
		require.NoError(t, err)

		bundle2, err := ephcli.GenerateAndEncryptAESKey(publicKeyPEM)
		require.NoError(t, err)

		// Each bundle should have different AES keys
		assert.NotEqual(t, bundle1.AESKey, bundle2.AESKey)
		assert.NotEqual(t, bundle1.HexString, bundle2.HexString)
		assert.NotEqual(t, bundle1.EncryptedAESKey, bundle2.EncryptedAESKey)
	})

	t.Run("error on invalid public key", func(t *testing.T) {
		t.Parallel()

		invalidPEM := "not a valid public key"

		bundle, err := ephcli.GenerateAndEncryptAESKey(invalidPEM)
		require.Error(t, err)
		assert.Nil(t, bundle)
	})
}

func TestSendAESKeyToEndpoint(t *testing.T) {
	t.Parallel()

	t.Run("successful send", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			// Verify request body
			var body map[string]string
			err := json.NewDecoder(r.Body).Decode(&body)
			require.NoError(t, err)
			assert.Contains(t, body, "aeskey")
			assert.Equal(t, "encrypted-key-123", body["aeskey"])

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status": "ok"}`))
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		err := client.SendAESKeyToEndpoint(server.URL+"/aeskey", "encrypted-key-123")
		require.NoError(t, err)
	})

	t.Run("error on non-200 status", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error": "invalid key"}`))
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		err := client.SendAESKeyToEndpoint(server.URL+"/aeskey", "invalid-key")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "400")
	})

	t.Run("error on server error", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		err := client.SendAESKeyToEndpoint(server.URL+"/aeskey", "some-key")
		require.Error(t, err)
	})

	t.Run("sends empty key", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body map[string]string
			err := json.NewDecoder(r.Body).Decode(&body)
			require.NoError(t, err)
			assert.Equal(t, "", body["aeskey"])

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		err := client.SendAESKeyToEndpoint(server.URL+"/aeskey", "")
		require.NoError(t, err)
	})
}

func TestE2EEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("AES key with special characters", func(t *testing.T) {
		t.Parallel()

		_, publicKeyPEM := generateTestRSAKeyPair(t)

		// Use hex string with all valid hex chars
		aesKeyHex := "0123456789abcdefABCDEF0123456789" +
			"0123456789abcdefABCDEF0123456789"

		encryptedKey, err := ephcli.EncryptAESKey(publicKeyPEM, aesKeyHex)
		require.NoError(t, err)
		assert.NotEmpty(t, encryptedKey)
	})

	t.Run("very long public key", func(t *testing.T) {
		t.Parallel()

		// Generate a 4096-bit key (longer)
		privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
		require.NoError(t, err)

		publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
		require.NoError(t, err)

		publicKeyPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: publicKeyBytes,
		})

		aesKey, err := ephcli.GenAESKey32bits()
		require.NoError(t, err)

		encryptedKey, err := ephcli.EncryptAESKey(string(publicKeyPEM), hex.EncodeToString(aesKey))
		require.NoError(t, err)
		assert.NotEmpty(t, encryptedKey)
	})

	t.Run("concurrent AES key generation", func(t *testing.T) {
		t.Parallel()

		// Generate multiple keys concurrently
		const numKeys = 10
		keys := make([][]byte, numKeys)
		errors := make([]error, numKeys)
		done := make(chan bool)

		for i := 0; i < numKeys; i++ {
			go func(index int) {
				keys[index], errors[index] = ephcli.GenAESKey32bits()
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numKeys; i++ {
			<-done
		}

		for i := 0; i < numKeys; i++ {
			require.NoError(t, errors[i])
			assert.Len(t, keys[i], 32)
		}

		// Verify all keys are unique
		keySet := make(map[string]bool)
		for _, key := range keys {
			keyStr := hex.EncodeToString(key)
			assert.False(t, keySet[keyStr], "Keys should be unique")
			keySet[keyStr] = true
		}
	})
}

func TestE2EWorkflow(t *testing.T) {
	t.Parallel()

	t.Run("complete E2E encryption workflow", func(t *testing.T) {
		t.Parallel()

		// Step 1: Generate RSA key pair
		_, publicKeyPEM := generateTestRSAKeyPair(t)

		// Step 2: Load public key
		pubKey, err := ephcli.LoadRSAPublicKey(publicKeyPEM)
		require.NoError(t, err)
		require.NotNil(t, pubKey)

		// Step 3: Generate and encrypt AES key
		bundle, err := ephcli.GenerateAndEncryptAESKey(publicKeyPEM)
		require.NoError(t, err)
		require.NotNil(t, bundle)

		// Verify bundle integrity
		assert.Len(t, bundle.AESKey, 32)
		assert.Equal(t, hex.EncodeToString(bundle.AESKey), bundle.HexString)
		assert.NotEmpty(t, bundle.EncryptedAESKey)

		// Step 4: Simulate sending to endpoint
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)

			var body map[string]string
			err := json.NewDecoder(r.Body).Decode(&body)
			require.NoError(t, err)
			assert.Equal(t, bundle.EncryptedAESKey, body["aeskey"])

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		err = client.SendAESKeyToEndpoint(server.URL, bundle.EncryptedAESKey)
		require.NoError(t, err)
	})
}
