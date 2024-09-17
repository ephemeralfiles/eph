package ephcli

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/schollz/progressbar/v3"
)

const (
	DefaultBarWidth = 50
)

func (c *ClientEphemeralfiles) DownloadEndpoint(uuidFile string) string {
	return fmt.Sprintf("%s/%s/download/%s", c.endpoint, apiVersion, uuidFile)
}

// Download downloads a file from the server
// and saves it to the outputfile
// If the outputfile is empty, the file will be saved to the current directory
// with the same name as the file on the server (retrieving the name from the Content-Disposition header)
func (c *ClientEphemeralfiles) Download(uuidFile string, outputfile string) error {
	var filename string
	url := c.DownloadEndpoint(uuidFile)
	ctx := context.Background()
	// prepare request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

	filename = c.getFileName(resp, outputfile)
	totalSize := resp.ContentLength

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer f.Close()

	if !c.noProgressBar {
		bar := progressbar.NewOptions64(totalSize,
			progressbar.OptionClearOnFinish(),
			progressbar.OptionShowBytes(true),
			progressbar.OptionSetWidth(DefaultBarWidth),
			progressbar.OptionSetDescription("downloading "+filename))

		_, err = io.Copy(io.MultiWriter(f, bar), resp.Body)
	} else {
		_, err = io.Copy(f, resp.Body)
	}
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}
	return nil
}

// getFileName returns outputFileName if not empty
// If empty, try to retrieve the filename from the Content-Disposition header
// If not present, return the last part of the URL
func (c *ClientEphemeralfiles) getFileName(resp *http.Response, outputfileName string) string {
	const defaultContentDispositionLength = 2
	var filename string

	if outputfileName != "" {
		return outputfileName
	}

	contentDisposition := resp.Header.Get("Content-Disposition")
	// extract filename from contentDisposition
	contentDispositionSplitted := strings.Split(contentDisposition, "=")
	if len(contentDispositionSplitted) < defaultContentDispositionLength {
		filename = filepath.Base(resp.Request.URL.Path)
	} else {
		filename = contentDispositionSplitted[1]
		// remove double quotes at the begin and the end from filename with regexp
		filename = strings.TrimLeft(strings.TrimRight(filename, "\""), "\"")
	}
	return filename
}