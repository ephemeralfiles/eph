package ephcli_test

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ephemeralfiles/eph/pkg/ephcli"
	"github.com/ephemeralfiles/eph/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "valid token",
			token: "test-token-123",
		},
		{
			name:  "empty token",
			token: "",
		},
		{
			name:  "long token",
			token: "very-long-token-" + fmt.Sprintf("%0*d", 100, 1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := ephcli.NewClient(tt.token)

			assert.NotNil(t, client)
		})
	}
}

func TestClientSetLogger(t *testing.T) {
	t.Parallel()

	client := ephcli.NewClient("test-token")
	testLogger := logger.NewLogger("debug")

	// Should not panic
	assert.NotPanics(t, func() {
		client.SetLogger(testLogger)
	})
}

func TestClientSetDebug(t *testing.T) {
	t.Parallel()

	client := ephcli.NewClient("test-token")

	// Should not panic
	assert.NotPanics(t, func() {
		client.SetDebug()
	})
}

func TestClientDisableProgressBar(t *testing.T) {
	t.Parallel()

	client := ephcli.NewClient("test-token")

	// Should not panic
	assert.NotPanics(t, func() {
		client.DisableProgressBar()
	})
}

func TestClientSetEndpoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		endpoint string
	}{
		{
			name:     "localhost endpoint",
			endpoint: "http://localhost:8080",
		},
		{
			name:     "production endpoint",
			endpoint: "https://api.example.com",
		},
		{
			name:     "endpoint with path",
			endpoint: "https://api.example.com/v1",
		},
		{
			name:     "empty endpoint",
			endpoint: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := ephcli.NewClient("test-token")

			// Should not panic
			assert.NotPanics(t, func() {
				client.SetEndpoint(tt.endpoint)
			})
		})
	}
}

func TestClientSetHTTPClient(t *testing.T) {
	t.Parallel()

	client := ephcli.NewClient("test-token")
	customHTTPClient := &http.Client{
		Timeout: 0,
	}

	// Should not panic
	assert.NotPanics(t, func() {
		client.SetHTTPClient(customHTTPClient)
	})
}

func TestClientHTTPOperations(t *testing.T) {
	t.Parallel()

	t.Run("successful request with auth header", func(t *testing.T) {
		t.Parallel()

		// Create a test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify Authorization header
			authHeader := r.Header.Get("Authorization")
			assert.Equal(t, "Bearer test-token", authHeader)

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[]`))
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		// Test using a public method that makes HTTP calls
		_, err := client.Fetch()
		// Should succeed with empty list
		require.NoError(t, err)
	})

	t.Run("error status code", func(t *testing.T) {
		t.Parallel()

		// Create a test server that returns error
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error": "unauthorized", "message": "invalid token"}`))
		}))
		defer server.Close()

		client := ephcli.NewClient("invalid-token")
		client.SetEndpoint(server.URL)

		_, err := client.Fetch()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "401")
	})

	t.Run("server error", func(t *testing.T) {
		t.Parallel()

		// Create a test server that returns 500
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error": "internal server error"}`))
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		_, err := client.Fetch()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "500")
	})
}

func TestClientWithCustomLogger(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		logLevel string
	}{
		{
			name:     "debug logger",
			logLevel: "debug",
		},
		{
			name:     "info logger",
			logLevel: "info",
		},
		{
			name:     "warn logger",
			logLevel: "warn",
		},
		{
			name:     "error logger",
			logLevel: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := ephcli.NewClient("test-token")
			testLogger := logger.NewLogger(tt.logLevel)

			assert.NotPanics(t, func() {
				client.SetLogger(testLogger)
			})
		})
	}
}

func TestClientWithNoLogger(t *testing.T) {
	t.Parallel()

	client := ephcli.NewClient("test-token")
	noLogger := logger.NoLogger()

	assert.NotPanics(t, func() {
		client.SetLogger(noLogger)
	})
}

func TestClientMultipleConfigurations(t *testing.T) {
	t.Parallel()

	// Test that multiple configurations can be applied
	client := ephcli.NewClient("test-token")

	assert.NotPanics(t, func() {
		client.SetEndpoint("http://localhost:8080")
		client.SetLogger(logger.NewLogger("debug"))
		client.SetDebug()
		client.DisableProgressBar()
		client.SetHTTPClient(&http.Client{})
	})
}

func TestClientEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("nil http client", func(t *testing.T) {
		t.Parallel()

		client := ephcli.NewClient("test-token")

		// SetHTTPClient with nil should not panic
		assert.NotPanics(t, func() {
			client.SetHTTPClient(nil)
		})
	})

	t.Run("nil logger", func(t *testing.T) {
		t.Parallel()

		client := ephcli.NewClient("test-token")

		// SetLogger with nil should not panic
		assert.NotPanics(t, func() {
			var nilLogger *slog.Logger
			client.SetLogger(nilLogger)
		})
	})

	t.Run("multiple set debug calls", func(t *testing.T) {
		t.Parallel()

		client := ephcli.NewClient("test-token")

		// Multiple SetDebug calls should not cause issues
		assert.NotPanics(t, func() {
			client.SetDebug()
			client.SetDebug()
			client.SetDebug()
		})
	})

	t.Run("multiple disable progress bar calls", func(t *testing.T) {
		t.Parallel()

		client := ephcli.NewClient("test-token")

		// Multiple DisableProgressBar calls should not cause issues
		assert.NotPanics(t, func() {
			client.DisableProgressBar()
			client.DisableProgressBar()
			client.DisableProgressBar()
		})
	})

	t.Run("various HTTP error responses", func(t *testing.T) {
		t.Parallel()

		errorTests := []struct {
			name       string
			statusCode int
			body       string
		}{
			{
				name:       "400 bad request",
				statusCode: http.StatusBadRequest,
				body:       `{"error": "bad request"}`,
			},
			{
				name:       "403 forbidden",
				statusCode: http.StatusForbidden,
				body:       `{"error": "forbidden"}`,
			},
			{
				name:       "404 not found",
				statusCode: http.StatusNotFound,
				body:       `{"error": "not found"}`,
			},
			{
				name:       "422 unprocessable entity",
				statusCode: http.StatusUnprocessableEntity,
				body:       `{"error": "validation failed"}`,
			},
			{
				name:       "429 rate limit",
				statusCode: http.StatusTooManyRequests,
				body:       `{"error": "rate limit exceeded"}`,
			},
			{
				name:       "502 bad gateway",
				statusCode: http.StatusBadGateway,
				body:       `{"error": "bad gateway"}`,
			},
			{
				name:       "503 service unavailable",
				statusCode: http.StatusServiceUnavailable,
				body:       `{"error": "service unavailable"}`,
			},
		}

		for _, tt := range errorTests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tt.statusCode)
					_, _ = w.Write([]byte(tt.body))
				}))
				defer server.Close()

				client := ephcli.NewClient("test-token")
				client.SetEndpoint(server.URL)

				_, err := client.Fetch()
				require.Error(t, err)
				assert.Contains(t, err.Error(), fmt.Sprintf("%d", tt.statusCode))
			})
		}
	})

	t.Run("plain text error response", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`Internal Server Error - Plain Text`))
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		_, err := client.Fetch()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "500")
	})

	t.Run("malformed JSON error response", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{invalid json`))
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		_, err := client.Fetch()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "400")
	})

	t.Run("empty error response body", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			// No body
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		_, err := client.Fetch()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "500")
	})
}
