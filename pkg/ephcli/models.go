package ephcli

import (
	"net/http"
	"time"
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
}

// NewClient creates a new client
func NewClient(token string) *ClientEphemeralfiles {
	return &ClientEphemeralfiles{
		httpClient:    &http.Client{},
		token:         token,
		endpoint:      defaultEndpoint,
		noProgressBar: false, // By default, the progress bar is active
	}
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
