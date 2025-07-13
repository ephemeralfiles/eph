package ephcli_test

import (
	"crypto/rand"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ephemeralfiles/eph/pkg/ephcli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownloadEndpoint(t *testing.T) {
	t.Parallel()
	// Create a new ClientEphemeralfiles
	client := ephcli.NewClient("token")
	client.SetEndpoint("https://ephemeralfiles.com")
	url := client.DownloadEndpoint("file1")
	assert.Equal(t, "https://ephemeralfiles.com/api/v1/download/file1", url)
}

func TestDownload(t *testing.T) {
	t.Parallel()
	t.Run("standard case: no error", func(t *testing.T) {
		t.Parallel()
		// Create a random file of 2MB
		f, err := os.Create("testfile")
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		// Write 2MB of random data
		_, err = io.CopyN(f, rand.Reader, 2*1024*1024)
		if err != nil {
			log.Fatal(err)
		}
		// Close the file
		f.Close()

		// Create a server
		ts := httptest.NewTLSServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check the method
				assert.Equal(t, http.MethodGet, r.Method)
				// Check the token
				assert.Equal(t, "Bearer token", r.Header.Get("Authorization"))
				// stream the file
				http.ServeFile(w, r, "testfile")
			}))
		defer ts.Close()
		client := ts.Client()

		// Create a new ClientEphemeralfiles
		e := ephcli.NewClient("token")
		e.SetHTTPClient(client)
		e.SetEndpoint(ts.URL)
		e.DisableProgressBar()

		// Download the file
		err = e.Download("file1", "testfile-downloaded")
		require.NoError(t, err)
		// calculate the checksum of the two files
		// checksum1, err := ephcli.Checksum("testfile")

		// Remove the file
		os.Remove("testfile")
		os.Remove("testfile-downloaded")
	})
}
