package ephcli_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ephemeralfiles/eph/pkg/dto"
	"github.com/ephemeralfiles/eph/pkg/ephcli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	t.Parallel()
	// TestList tests the List function
	t.Run("standard case: no error", func(t *testing.T) {
		t.Parallel()
		// Simulate a server
		response := dto.FileList{
			dto.File{
				FileID:   "1",
				FileName: "file1",
				Size:     100,
			},
			{
				FileID:   "2",
				FileName: "file2",
				Size:     200,
			},
		}
		responseJSON, _ := json.Marshal(response)
		ts := httptest.NewTLSServer(
			http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				fmt.Fprintln(w, string(responseJSON))
			}))

		defer ts.Close()
		client := ts.Client()

		e := ephcli.NewClient("token")
		e.SetHTTPClient(client)
		e.SetEndpoint(ts.URL)

		files, err := e.Fetch()
		require.NoError(t, err)
		assert.Len(t, files, 2)
	})
	t.Run("Check token is in header and method is GET", func(t *testing.T) {
		t.Parallel()
		// Simulate a server
		response := dto.FileList{}
		responseJSON, _ := json.Marshal(response)
		ts := httptest.NewTLSServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "Bearer token", r.Header.Get("Authorization"))
				assert.Equal(t, http.MethodGet, r.Method)
				fmt.Fprintln(w, string(responseJSON))
			}))

		defer ts.Close()

		client := ts.Client()

		e := ephcli.NewClient("token")
		e.SetHTTPClient(client)
		e.SetEndpoint(ts.URL)

		_, err := e.Fetch()
		require.NoError(t, err)
	})

	t.Run("Check List handles errors correctly if response is not well formatted", func(t *testing.T) {
		t.Parallel()
		ts := httptest.NewTLSServer(
			http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				fmt.Fprintln(w, "not a json")
			}))
		defer ts.Close()

		client := ts.Client()

		e := ephcli.NewClient("token")
		e.SetHTTPClient(client)
		e.SetEndpoint(ts.URL)

		_, err := e.Fetch()
		assert.Error(t, err)
	})
	t.Run("Check List handles errors correctly if http status code is not 200", func(t *testing.T) {
		t.Parallel()
		ts := httptest.NewTLSServer(
			http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintln(w, "{}")
			}))
		defer ts.Close()

		client := ts.Client()

		e := ephcli.NewClient("token")
		e.SetHTTPClient(client)
		e.SetEndpoint(ts.URL)

		_, err := e.Fetch()
		require.Error(t, err)
	})
}

func TestPrint(t *testing.T) {
	t.Parallel()
	t.Run("case with no data", func(t *testing.T) {
		t.Parallel()
		fl := dto.FileList{}
		// table
		err := ephcli.Print(&fl)
		require.NoError(t, err)
		// CSV
		err = ephcli.PrintCSV(&fl)
		require.NoError(t, err)
		// JSON
		err = ephcli.PrintJSON(&fl)
		require.NoError(t, err)
		// YAML
		err = ephcli.PrintYAML(&fl)
		require.NoError(t, err)
	})
	t.Run("case with data", func(t *testing.T) {
		t.Parallel()
		fl := dto.FileList{
			dto.File{
				FileID:   "1",
				FileName: "file1",
				Size:     100,
			},
			{
				FileID:   "2",
				FileName: "file2",
				Size:     200,
			},
		}
		// table
		err := ephcli.Print(&fl)
		require.NoError(t, err)
		// CSV
		err = ephcli.PrintCSV(&fl)
		require.NoError(t, err)
		// JSON
		err = ephcli.PrintJSON(&fl)
		require.NoError(t, err)
		// YAML
		err = ephcli.PrintYAML(&fl)
		require.NoError(t, err)
	})
}
