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
	DefaultAPIRequestTimeout = 5 * time.Second
)

var (
	ErrGettingLatestRelease = errors.New("error getting latest release")
)

type Client struct {
	httpClient *http.Client
	endpoint   string
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{},
		endpoint:   DefaultEndpoint,
	}
}

func (s *Client) SetHTTPClient(client *http.Client) {
	s.httpClient = client
}

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
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error getting latest release: %s", resp.Status) //nolint:goerr113
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
