package ephcli_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ephemeralfiles/eph/pkg/ephcli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemove(t *testing.T) {
	t.Parallel()
	t.Run("standard case: no error", func(t *testing.T) {
		t.Parallel()
		ts := httptest.NewTLSServer(
			http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

		defer ts.Close()
		client := ts.Client()

		e := ephcli.NewClient("token")
		e.SetHTTPClient(client)
		e.SetEndpoint(ts.URL)

		err := e.Remove("uuid")
		require.NoError(t, err)
	})
	t.Run("Check token is in header and method is DELETE", func(t *testing.T) {
		t.Parallel()
		// Simulate a server
		ts := httptest.NewTLSServer(
			http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "Bearer token", r.Header.Get("Authorization"))
				assert.Equal(t, "DELETE", r.Method)
			}))

		defer ts.Close()

		client := ts.Client()

		e := ephcli.NewClient("token")
		e.SetHTTPClient(client)
		e.SetEndpoint(ts.URL)

		err := e.Remove("uuid")
		require.NoError(t, err)
	})

	t.Run("Check that errors correctly if response is not well formatted", func(t *testing.T) {
		t.Parallel()
		ts := httptest.NewTLSServer(
			http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintln(w, "not a json")
			}))
		defer ts.Close()

		client := ts.Client()

		e := ephcli.NewClient("token")
		e.SetHTTPClient(client)
		e.SetEndpoint(ts.URL)

		err := e.Remove("")
		assert.Error(t, err)
	})
	t.Run("Check Remove handles errors correctly if http status code is not 200", func(t *testing.T) {
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

		err := e.Remove("")
		require.Error(t, err)
	})
}
