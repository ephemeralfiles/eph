package ephcli

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/ephemeralfiles/eph/pkg/logger"
	"github.com/schollz/progressbar/v3"
)

const DefaultAPIRequestTimeout = 5 * time.Second

// apiVersion is the version of the API that the client expects
const apiVersion string = "api/v1"

// defaultEndpoint is the default endpoint of the API
const defaultEndpoint string = "https://ephemeralfiles.com"

// ClientEphemeralfiles is the client to interact with the API
type ClientEphemeralfiles struct {
	httpClient    *http.Client
	token         string
	endpoint      string
	noProgressBar bool
	bar           *progressbar.ProgressBar
	log           *slog.Logger
}

// NewClient creates a new client
func NewClient(token string) *ClientEphemeralfiles {
	return &ClientEphemeralfiles{
		httpClient:    &http.Client{},
		token:         token,
		endpoint:      defaultEndpoint,
		noProgressBar: false, // By default, the progress bar is active
		log:           logger.NoLogger(),
	}
}

// SetLogger sets the logger
func (c *ClientEphemeralfiles) SetLogger(logger *slog.Logger) {
	c.log = logger
}

// SetDebug sets the logger to debug
// Disable the progress bar
func (c *ClientEphemeralfiles) SetDebug() {
	c.log = logger.NewLogger("debug")
	c.noProgressBar = true
}

func (c *ClientEphemeralfiles) DisableProgressBar() {
	c.noProgressBar = true
}

func (c *ClientEphemeralfiles) SetEndpoint(endpoint string) {
	c.endpoint = endpoint
}

func (c *ClientEphemeralfiles) SetHTTPClient(client *http.Client) {
	c.httpClient = client
}

// HTTP utility methods to reduce duplication

// addAuthHeader adds the Bearer token to the request
func (c *ClientEphemeralfiles) addAuthHeader(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.token)
}

// checkResponseStatus checks if the response status is OK and handles errors
func (c *ClientEphemeralfiles) checkResponseStatus(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return parseError(resp)
	}
	return nil
}

// createRequestWithTimeout creates an HTTP request with the default timeout context
func (c *ClientEphemeralfiles) createRequestWithTimeout(method, url string, body io.Reader) (*http.Request, context.CancelFunc, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultAPIRequestTimeout)
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		cancel()
		return nil, nil, err
	}
	return req, cancel, nil
}

// doRequestWithAuth performs an HTTP request with authentication and status checking
func (c *ClientEphemeralfiles) doRequestWithAuth(req *http.Request) (*http.Response, error) {
	c.addAuthHeader(req)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	
	if err := c.checkResponseStatus(resp); err != nil {
		resp.Body.Close()
		return nil, err
	}
	
	return resp, nil
}
