package ephcli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Box represents storage capacity and usage information for a user's box.
type Box struct {
	CapacityMb  int64 `json:"capacity_mb"`
	UsedMb      int64 `json:"used_mb"`
	RemainingMb int64 `json:"remaining_mb"`
}

// GetBoxInfos retrieves storage capacity and usage information for the current user's box.
func (c *ClientEphemeralfiles) GetBoxInfos() (*Box, error) {
	var b Box

	email, _, err := Whoami(c.token)
	if err != nil {
		return nil, fmt.Errorf("error getting user info: %w", err)
	}

	url := fmt.Sprintf("%s/%s/box/%s/default", c.endpoint, apiVersion, email)
	ctx, cancel := context.WithTimeout(context.Background(), DefaultAPIRequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.httpClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, parseError(resp)
	}

	err = json.NewDecoder(resp.Body).Decode(&b)
	if err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}
	return &b, nil
}
