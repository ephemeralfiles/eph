package ephcli

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/schollz/progressbar/v3"
)

func (c *ClientEphemeralfiles) UploadEndpoint() string {
	return fmt.Sprintf("%s/%s/upload", c.endpoint, apiVersion)
}

func (c *ClientEphemeralfiles) Upload(fileToUpload string) error {
	stat, err := os.Stat(fileToUpload)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: %s", ErrFileNotFound, fileToUpload)
		}
		return fmt.Errorf("error getting file info: %w", err)
	}

	bar := progressbar.NewOptions64(stat.Size(), progressbar.OptionClearOnFinish(),
		progressbar.OptionShowBytes(true), progressbar.OptionSetWidth(DefaultBarWidth),
		progressbar.OptionSetDescription("uploadding file..."),
		progressbar.OptionSetVisibility(!c.noProgressBar),
	)

	// Create a multipart form
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go func() {
		defer pw.Close()
		part, err := writer.CreateFormFile("uploadfile", fileToUpload)
		if err != nil {
			pw.CloseWithError(err) // Properly handle the error
			return
		}
		f, err := os.Open(fileToUpload)
		if err != nil {
			pw.CloseWithError(err) // Properly handle the error
			return
		}
		defer f.Close()
		if _, err := io.Copy(io.MultiWriter(part, bar), f); err != nil {
			pw.CloseWithError(err) // Properly handle the error
			return
		}
		writer.Close()
	}()

	ctx := context.Background()
	// Create a request with the form
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.UploadEndpoint(), pr)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	// Add the token to the request
	req.Header.Add("Authorization", "Bearer "+c.token)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		_ = bar.RenderBlank()
		return parseError(resp)
	}
	return nil
}
