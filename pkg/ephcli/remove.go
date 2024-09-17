package ephcli

import (
	"context"
	"fmt"
	"net/http"
)

// Download downloads a file from the server
// and saves it to the outputfile
// If the outputfile is empty, the file will be saved to the current directory
// with the same name as the file on the server (retrieving the name from the Content-Disposition header)
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return parseError(resp)
	}
	defer resp.Body.Close()
	return nil
}
