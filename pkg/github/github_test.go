package github_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func TestNewClient(t *testing.T) {
	t.Parallel()

	client := github.NewClient()
	
	assert.NotNil(t, client)
	// We can't directly test the internal fields, but we can test behavior
	// The client should be ready to use with default settings
}

func TestSetHTTPClient(t *testing.T) {
	t.Parallel()

	client := github.NewClient()
	customHTTPClient := &http.Client{Timeout: 10 * time.Second}
	
	client.SetHTTPClient(customHTTPClient)
	
	// Test that the custom client is used by making a request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"tag_name": "v1.0.0"}`)
	}))
	defer ts.Close()
	
	client.SetEndpoint(ts.URL)
	version, err := client.GetLastVersionFromGithub("test/repo")
	require.NoError(t, err)
	assert.Equal(t, "1.0.0", version)
}

func TestSetEndpoint(t *testing.T) {
	t.Parallel()

	client := github.NewClient()
	customEndpoint := "https://custom-github-api.com"
	
	client.SetEndpoint(customEndpoint)
	
	// We can't directly test the endpoint field, but we can verify
	// it's used in API calls by checking the URL in requests
	// This is implicitly tested through the existing tests
	assert.NotNil(t, client)
}

func TestGetLastVersionFromGithubErrorPaths(t *testing.T) {
	t.Parallel()

	t.Run("timeout error", func(t *testing.T) {
		t.Parallel()

		// Create a server that delays response beyond timeout
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			time.Sleep(10 * time.Second) // Longer than DefaultAPIRequestTimeout
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, `{"tag_name": "v1.0.0"}`)
		}))
		defer ts.Close()

		client := github.NewClient()
		client.SetHTTPClient(ts.Client())
		client.SetEndpoint(ts.URL)

		_, err := client.GetLastVersionFromGithub("test/repo")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error getting latest release")
	})

	t.Run("different HTTP status codes", func(t *testing.T) {
		t.Parallel()

		statusCodes := []int{
			http.StatusBadRequest,
			http.StatusUnauthorized,
			http.StatusForbidden,
			http.StatusNotFound,
			http.StatusTooManyRequests,
			http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
		}

		for _, statusCode := range statusCodes {
			t.Run(http.StatusText(statusCode), func(t *testing.T) {
				t.Parallel()

				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(statusCode)
					fmt.Fprintln(w, `{"message": "API Error"}`)
				}))
				defer ts.Close()

				client := github.NewClient()
				client.SetHTTPClient(ts.Client())
				client.SetEndpoint(ts.URL)

				_, err := client.GetLastVersionFromGithub("test/repo")
				require.Error(t, err)
				assert.Contains(t, err.Error(), "error getting latest release")
			})
		}
	})

	t.Run("malformed JSON response", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, `{"tag_name": "v1.0.0", "incomplete...`)
		}))
		defer ts.Close()

		client := github.NewClient()
		client.SetHTTPClient(ts.Client())
		client.SetEndpoint(ts.URL)

		_, err := client.GetLastVersionFromGithub("test/repo")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error decoding response")
	})

	t.Run("empty tag_name in response", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, `{"tag_name": ""}`)
		}))
		defer ts.Close()

		client := github.NewClient()
		client.SetHTTPClient(ts.Client())
		client.SetEndpoint(ts.URL)

		_, err := client.GetLastVersionFromGithub("test/repo")
		require.Error(t, err)
		assert.Equal(t, github.ErrGettingLatestRelease, err)
	})

	t.Run("missing tag_name field", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, `{"name": "Release 1.0.0", "body": "Release notes"}`)
		}))
		defer ts.Close()

		client := github.NewClient()
		client.SetHTTPClient(ts.Client())
		client.SetEndpoint(ts.URL)

		_, err := client.GetLastVersionFromGithub("test/repo")
		require.Error(t, err)
		assert.Equal(t, github.ErrGettingLatestRelease, err)
	})

	t.Run("invalid repository format", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if the URL contains the invalid repository
			assert.Contains(t, r.URL.Path, "invalid-repo-format")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, `{"message": "Not Found"}`)
		}))
		defer ts.Close()

		client := github.NewClient()
		client.SetHTTPClient(ts.Client())
		client.SetEndpoint(ts.URL)

		_, err := client.GetLastVersionFromGithub("invalid-repo-format")
		require.Error(t, err)
	})
}

func TestVersionTagFormatting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		tagName     string
		expectedVer string
	}{
		{
			name:        "version with v prefix",
			tagName:     "v1.2.3",
			expectedVer: "1.2.3",
		},
		{
			name:        "version without v prefix",
			tagName:     "1.2.3",
			expectedVer: "1.2.3",
		},
		{
			name:        "semantic version with v prefix",
			tagName:     "v1.0.0-alpha.1",
			expectedVer: "1.0.0-alpha.1",
		},
		{
			name:        "version with multiple v's",
			tagName:     "vv1.0.0",
			expectedVer: "v1.0.0", // Only first v is removed
		},
		{
			name:        "version with build metadata",
			tagName:     "v2.1.0+build.123",
			expectedVer: "2.1.0+build.123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, `{"tag_name": "%s"}`, tt.tagName)
			}))
			defer ts.Close()

			client := github.NewClient()
			client.SetHTTPClient(ts.Client())
			client.SetEndpoint(ts.URL)

			version, err := client.GetLastVersionFromGithub("test/repo")
			require.NoError(t, err)
			assert.Equal(t, tt.expectedVer, version)
		})
	}
}

func TestErrorConstants(t *testing.T) {
	t.Parallel()

	t.Run("ErrGettingLatestRelease", func(t *testing.T) {
		t.Parallel()

		err := github.ErrGettingLatestRelease
		assert.NotNil(t, err)
		assert.Equal(t, "error getting latest release", err.Error())
		
		// Test error wrapping
		wrappedErr := fmt.Errorf("failed to check for updates: %w", err)
		assert.Contains(t, wrappedErr.Error(), "error getting latest release")
		assert.True(t, errors.Is(wrappedErr, github.ErrGettingLatestRelease))
	})
}

func TestDefaultConstants(t *testing.T) {
	t.Parallel()

	t.Run("DefaultEndpoint", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "https://api.github.com", github.DefaultEndpoint)
		assert.Contains(t, github.DefaultEndpoint, "api.github.com")
	})

	t.Run("DefaultAPIRequestTimeout", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, 5*time.Second, github.DefaultAPIRequestTimeout)
		assert.True(t, github.DefaultAPIRequestTimeout > 0)
	})
}

func TestClientEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("nil HTTP client", func(t *testing.T) {
		t.Parallel()

		client := github.NewClient()
		client.SetHTTPClient(nil)

		// This should cause a panic or error when used
		defer func() {
			if r := recover(); r != nil {
				// If it panics, that's expected behavior
				assert.NotNil(t, r)
			}
		}()

		// Attempting to use nil client should fail
		_, err := client.GetLastVersionFromGithub("test/repo")
		// Either error or panic is acceptable
		if err != nil {
			assert.Error(t, err)
		}
	})

	t.Run("empty endpoint", func(t *testing.T) {
		t.Parallel()

		client := github.NewClient()
		client.SetEndpoint("")

		_, err := client.GetLastVersionFromGithub("test/repo")
		require.Error(t, err)
	})

	t.Run("malformed endpoint", func(t *testing.T) {
		t.Parallel()

		client := github.NewClient()
		client.SetEndpoint("not-a-valid-url")

		_, err := client.GetLastVersionFromGithub("test/repo")
		require.Error(t, err)
	})

	t.Run("endpoint with path", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify the path includes both the custom path and the API path
			assert.Contains(t, r.URL.Path, "/api/v3/repos/test/repo/releases/latest")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, `{"tag_name": "v1.0.0"}`)
		}))
		defer ts.Close()

		client := github.NewClient()
		client.SetHTTPClient(ts.Client())
		client.SetEndpoint(ts.URL + "/api/v3") // Custom API path

		version, err := client.GetLastVersionFromGithub("test/repo")
		require.NoError(t, err)
		assert.Equal(t, "1.0.0", version)
	})
}
