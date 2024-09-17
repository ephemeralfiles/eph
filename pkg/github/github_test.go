package github_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ephemeralfiles/eph/pkg/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetLastVersionFromGithubs(t *testing.T) {
	t.Parallel()
	t.Run("standard case, no error", func(t *testing.T) {
		t.Parallel()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, `{"tag_name": "v0.1.0"}`)
		}))
		defer ts.Close()

		client := ts.Client()

		g := github.NewClient()
		g.SetHTTPClient(client)
		g.SetEndpoint(ts.URL)

		b, err := g.GetLastVersionFromGithub("ephemeralfiles/eph")
		require.NoError(t, err)
		assert.NotNil(t, b)
		assert.Equal(t, "0.1.0", b)
	})

	t.Run("case with no data", func(t *testing.T) {
		t.Parallel()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, `{}`)
		}))
		defer ts.Close()

		client := ts.Client()

		g := github.NewClient()
		g.SetHTTPClient(client)
		g.SetEndpoint(ts.URL)

		b, err := g.GetLastVersionFromGithub("ephemeralfiles/eph")
		require.Error(t, err)
		assert.Equal(t, "", b)
	})

	t.Run("case with error", func(t *testing.T) {
		t.Parallel()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "{}")
		}))
		defer ts.Close()

		client := ts.Client()

		g := github.NewClient()
		g.SetHTTPClient(client)
		g.SetEndpoint(ts.URL)

		_, err := g.GetLastVersionFromGithub("ephemeralfiles/eph")
		require.Error(t, err)
	})

	t.Run("case when server do not respond", func(t *testing.T) {
		t.Parallel()
		g := github.NewClient()
		g.SetEndpoint("http://localhost:9999/do-not-exist")
		_, err := g.GetLastVersionFromGithub("ephemeralfiles/eph")
		require.Error(t, err)
	})

	t.Run("case with wrong data returned", func(t *testing.T) {
		t.Parallel()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "lskdfkjbsd")
		}))
		defer ts.Close()

		client := ts.Client()

		g := github.NewClient()
		g.SetHTTPClient(client)
		g.SetEndpoint(ts.URL)

		_, err := g.GetLastVersionFromGithub("ephemeralfiles/eph")
		require.Error(t, err)
	})
}
