package ephcli

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/ephemeralfiles/eph/pkg/logger"
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
