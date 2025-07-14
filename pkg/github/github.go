// Package github provides functionality for interacting with GitHub releases
// and checking for updates to the ephemeral files CLI.
package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	// DefaultEndpoint is the default GitHub API endpoint.
	DefaultEndpoint          = "https://api.github.com"
	// DefaultAPIRequestTimeout is the default timeout for GitHub API requests.
	DefaultAPIRequestTimeout = 5 * time.Second
)

var (
	// ErrGettingLatestRelease is returned when failing to fetch the latest release from GitHub.
	ErrGettingLatestRelease = errors.New("error getting latest release")
)

// Client represents a GitHub API client for fetching release information.
type Client struct {
	httpClient *http.Client
	endpoint   string
}

// NewClient creates a new GitHub API client with default settings.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{},
		endpoint:   DefaultEndpoint,
	}
}

// SetHTTPClient sets the HTTP client for the GitHub client.
func (s *Client) SetHTTPClient(client *http.Client) {
	s.httpClient = client
}

// SetEndpoint sets the GitHub API endpoint for the client.
func (s *Client) SetEndpoint(endpoint string) {
	s.endpoint = endpoint
}

// GetLastVersionFromGithub gets the last version from github.
func (s *Client) GetLastVersionFromGithub(repository string) (string, error) {
	var (
		release ResponseLatestRelease
		resp    *http.Response
		err     error
	)
	url := fmt.Sprintf("%s/repos/%s/releases/latest", s.endpoint, repository)
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(DefaultAPIRequestTimeout))
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("error getting latest release: %w", err)
	}
	resp, err = s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error getting latest release: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: %s", ErrGettingLatestRelease, resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		return "", fmt.Errorf("error decoding response: %w: %s", err, resp.Body)
	}

	if release.TagName == "" {
		return "", ErrGettingLatestRelease
	}
	// remove the v prefix
	tag := strings.TrimPrefix(release.TagName, "v")
	return tag, nil
}
