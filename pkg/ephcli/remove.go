package ephcli

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
)

// Remove deletes a file from the ephemeralfiles service by its UUID.
func (c *ClientEphemeralfiles) Remove(uuidFileToRemove string) error {
	url := fmt.Sprintf("%s/%s", c.FilesEndpoint(), uuidFileToRemove)
	ctx, cancel := context.WithTimeout(context.Background(), DefaultAPIRequestTimeout)
	defer cancel()

	// prepare request
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.log.Debug("Warning: failed to close response body", slog.String("error", closeErr.Error()))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return parseError(resp)
	}
	return nil
}
