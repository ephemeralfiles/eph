package ephcli_test

import (
	"crypto/rand"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/ephemeralfiles/eph/pkg/ephcli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUploadEndpoint(t *testing.T) {
	t.Parallel()

	client := ephcli.NewClient("test-token")
	client.SetEndpoint("https://test.ephemeralfiles.com")
	
	endpoint := client.UploadEndpoint()
	expected := "https://test.ephemeralfiles.com/api/v1/upload/clear"
	assert.Equal(t, expected, endpoint)
}

func TestUploadEndpointDefaultEndpoint(t *testing.T) {
	t.Parallel()

	client := ephcli.NewClient("test-token")
	// Don't set endpoint, use default
	
	endpoint := client.UploadEndpoint()
	expected := "https://ephemeralfiles.com/api/v1/upload/clear"
	assert.Equal(t, expected, endpoint)
}

func TestUpload(t *testing.T) {
	t.Parallel()

	t.Run("successful upload", func(t *testing.T) {
		t.Parallel()

		// Create a temporary test file
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "test-upload.txt")
		testContent := "This is test content for upload"
		
		err := os.WriteFile(testFile, []byte(testContent), 0644)
		require.NoError(t, err)

		// Create a mock server
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request method
			assert.Equal(t, http.MethodPost, r.Method)
			
			// Verify authorization header
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
			
			// Verify content type is multipart/form-data
			contentType := r.Header.Get("Content-Type")
			assert.Contains(t, contentType, "multipart/form-data")
			
			// Parse multipart form
			err := r.ParseMultipartForm(32 << 20) // 32MB
			require.NoError(t, err)
			
			// Verify file was uploaded
			file, fileHeader, err := r.FormFile("uploadfile")
			require.NoError(t, err)
			defer file.Close()
			
			assert.Equal(t, filepath.Base(testFile), fileHeader.Filename)
			
			// Read uploaded content and verify
			uploadedContent, err := io.ReadAll(file)
			require.NoError(t, err)
			assert.Equal(t, testContent, string(uploadedContent))
			
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		// Create client with test server
		client := ephcli.NewClient("test-token")
		client.SetEndpoint(ts.URL)
		client.DisableProgressBar()

		// Perform upload
		err = client.Upload(testFile)
		assert.NoError(t, err)
	})

	t.Run("file not found error", func(t *testing.T) {
		t.Parallel()

		client := ephcli.NewClient("test-token")
		client.DisableProgressBar()

		err := client.Upload("nonexistent-file.txt")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nonexistent-file.txt")
	})

	t.Run("server error response", func(t *testing.T) {
		t.Parallel()

		// Create a temporary test file
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "test-error.txt")
		err := os.WriteFile(testFile, []byte("test"), 0644)
		require.NoError(t, err)

		// Create a mock server that returns an error
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": true, "msg": "Server error occurred"}`))
		}))
		defer ts.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(ts.URL)
		client.DisableProgressBar()

		err = client.Upload(testFile)
		assert.Error(t, err)
	})

	t.Run("large file upload", func(t *testing.T) {
		t.Parallel()

		// Create a larger temporary test file (1MB)
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "large-test.bin")
		
		f, err := os.Create(testFile)
		require.NoError(t, err)
		defer f.Close()
		
		// Write 1MB of random data
		_, err = io.CopyN(f, rand.Reader, 1024*1024)
		require.NoError(t, err)
		f.Close()

		var receivedSize int64
		
		// Create a mock server
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
			
			// Parse multipart form
			err := r.ParseMultipartForm(32 << 20)
			require.NoError(t, err)
			
			file, _, err := r.FormFile("uploadfile")
			require.NoError(t, err)
			defer file.Close()
			
			// Count bytes received
			receivedSize, err = io.Copy(io.Discard, file)
			require.NoError(t, err)
			
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(ts.URL)
		client.DisableProgressBar()

		err = client.Upload(testFile)
		assert.NoError(t, err)
		assert.Equal(t, int64(1024*1024), receivedSize)
	})

	t.Run("network error", func(t *testing.T) {
		t.Parallel()

		// Create a temporary test file
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "test-network.txt")
		err := os.WriteFile(testFile, []byte("test"), 0644)
		require.NoError(t, err)

		client := ephcli.NewClient("test-token")
		// Set invalid endpoint to trigger network error
		client.SetEndpoint("http://invalid-host-that-does-not-exist.local")
		client.DisableProgressBar()

		err = client.Upload(testFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error sending request")
	})
}

func TestValidateAndGetFileInfo(t *testing.T) {
	t.Parallel()

	// We need to test this indirectly since it's not exported
	// We'll test through the Upload method behavior

	t.Run("valid file", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "valid-file.txt")
		testContent := "valid content"
		
		err := os.WriteFile(testFile, []byte(testContent), 0644)
		require.NoError(t, err)

		// Create a mock server that accepts the upload
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(ts.URL)
		client.DisableProgressBar()

		// This should succeed, indicating file validation passed
		err = client.Upload(testFile)
		assert.NoError(t, err)
	})

	t.Run("nonexistent file", func(t *testing.T) {
		t.Parallel()

		client := ephcli.NewClient("test-token")
		client.DisableProgressBar()

		err := client.Upload("/path/that/does/not/exist.txt")
		assert.Error(t, err)
		// The error should mention the file not being found
		assert.Contains(t, err.Error(), "/path/that/does/not/exist.txt")
	})

	t.Run("directory instead of file", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()

		client := ephcli.NewClient("test-token")
		client.DisableProgressBar()

		err := client.Upload(tempDir)
		assert.Error(t, err)
	})

	t.Run("permission denied file", func(t *testing.T) {
		t.Parallel()

		// This test might not work on all systems, so we'll skip it if needed
		if os.Getenv("CI") != "" {
			t.Skip("Skipping permission test in CI environment")
		}

		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "no-permission.txt")
		
		err := os.WriteFile(testFile, []byte("test"), 0644)
		require.NoError(t, err)
		
		// Remove read permissions
		err = os.Chmod(testFile, 0000)
		require.NoError(t, err)
		
		// Restore permissions after test
		defer func() {
			os.Chmod(testFile, 0644)
		}()

		client := ephcli.NewClient("test-token")
		client.DisableProgressBar()

		err = client.Upload(testFile)
		assert.Error(t, err)
	})
}

func TestCreateMultipartForm(t *testing.T) {
	t.Parallel()

	// Test multipart form creation indirectly through upload behavior
	t.Run("multipart form structure", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "multipart-test.txt")
		testContent := "multipart form test content"
		
		err := os.WriteFile(testFile, []byte(testContent), 0644)
		require.NoError(t, err)

		var receivedFormData map[string]string
		var receivedFile []byte
		var receivedFilename string

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Parse the multipart form
			err := r.ParseMultipartForm(32 << 20)
			require.NoError(t, err)

			receivedFormData = make(map[string]string)
			for key, values := range r.MultipartForm.Value {
				if len(values) > 0 {
					receivedFormData[key] = values[0]
				}
			}

			// Get the uploaded file
			file, fileHeader, err := r.FormFile("uploadfile")
			require.NoError(t, err)
			defer file.Close()

			receivedFilename = fileHeader.Filename
			receivedFile, err = io.ReadAll(file)
			require.NoError(t, err)

			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(ts.URL)
		client.DisableProgressBar()

		err = client.Upload(testFile)
		require.NoError(t, err)

		// Verify the multipart form was created correctly
		assert.Equal(t, filepath.Base(testFile), receivedFilename)
		assert.Equal(t, testContent, string(receivedFile))
	})
}

func TestSendUploadRequest(t *testing.T) {
	t.Parallel()

	// Test request sending indirectly through upload behavior
	t.Run("request headers and method", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "headers-test.txt")
		err := os.WriteFile(testFile, []byte("test"), 0644)
		require.NoError(t, err)

		var receivedHeaders http.Header
		var receivedMethod string

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedHeaders = r.Header.Clone()
			receivedMethod = r.Method
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		client := ephcli.NewClient("custom-token-123")
		client.SetHTTPClient(ts.Client())
		client.SetEndpoint(ts.URL)
		client.DisableProgressBar()

		err = client.Upload(testFile)
		require.NoError(t, err)

		// Verify request properties
		assert.Equal(t, http.MethodPost, receivedMethod)
		assert.Equal(t, "Bearer custom-token-123", receivedHeaders.Get("Authorization"))
		assert.Contains(t, receivedHeaders.Get("Content-Type"), "multipart/form-data")
		assert.Contains(t, receivedHeaders.Get("Content-Type"), "boundary=")
	})

	t.Run("response status code handling", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "status-test.txt")
		err := os.WriteFile(testFile, []byte("test"), 0644)
		require.NoError(t, err)

		statusCodes := []int{
			http.StatusOK,
			http.StatusBadRequest,
			http.StatusUnauthorized,
			http.StatusForbidden,
			http.StatusNotFound,
			http.StatusInternalServerError,
		}

		for _, statusCode := range statusCodes {
			t.Run(http.StatusText(statusCode), func(t *testing.T) {
				t.Parallel()

				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(statusCode)
					if statusCode != http.StatusOK {
						w.Write([]byte(`{"error": true, "msg": "Error message"}`))
					}
				}))
				defer ts.Close()

				client := ephcli.NewClient("test-token")
				client.SetHTTPClient(ts.Client())
				client.SetEndpoint(ts.URL)
				client.DisableProgressBar()

				err = client.Upload(testFile)
				
				if statusCode == http.StatusOK {
					assert.NoError(t, err)
				} else {
					assert.Error(t, err)
				}
			})
		}
	})
}

func TestUploadEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("empty file", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "empty.txt")
		
		// Create empty file
		f, err := os.Create(testFile)
		require.NoError(t, err)
		f.Close()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseMultipartForm(32 << 20)
			require.NoError(t, err)
			
			file, _, err := r.FormFile("uploadfile")
			require.NoError(t, err)
			defer file.Close()
			
			content, err := io.ReadAll(file)
			require.NoError(t, err)
			assert.Empty(t, content)
			
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(ts.URL)
		client.DisableProgressBar()

		err = client.Upload(testFile)
		assert.NoError(t, err)
	})

	t.Run("file with special characters in name", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		// Create file with special characters
		testFile := filepath.Join(tempDir, "test file with spaces & symbols!@#.txt")
		err := os.WriteFile(testFile, []byte("content"), 0644)
		require.NoError(t, err)

		var receivedFilename string

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseMultipartForm(32 << 20)
			require.NoError(t, err)
			
			_, fileHeader, err := r.FormFile("uploadfile")
			require.NoError(t, err)
			
			receivedFilename = fileHeader.Filename
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(ts.URL)
		client.DisableProgressBar()

		err = client.Upload(testFile)
		require.NoError(t, err)
		assert.Equal(t, filepath.Base(testFile), receivedFilename)
	})

	t.Run("upload with progress bar enabled", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "progress-test.txt")
		err := os.WriteFile(testFile, []byte("progress test content"), 0644)
		require.NoError(t, err)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		client := ephcli.NewClient("test-token")
		client.SetHTTPClient(ts.Client())
		client.SetEndpoint(ts.URL)
		// Progress bar enabled (default)

		err = client.Upload(testFile)
		assert.NoError(t, err)
	})
}