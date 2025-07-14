package ephcli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/ephemeralfiles/eph/pkg/dto"
)

var (
	// ErrFileNotFound is returned when a requested file cannot be found.
	ErrFileNotFound = errors.New("file not found")
)

// parseError is a helper function to parse the error from the response.
func parseError(resp *http.Response) error {
	var jsonResponse dto.APIError
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	err = json.Unmarshal(respBody, &jsonResponse)
	if err != nil {
		return fmt.Errorf("error decoding response: %w - raw response: %s", err, respBody)
	}
	return fmt.Errorf("status not ok %d: %s", resp.StatusCode, jsonResponse.Message) //nolint:err113
}
