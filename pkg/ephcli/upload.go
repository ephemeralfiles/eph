package ephcli

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
)

// UploadEndpoint returns the API endpoint URL for file uploads.
func (c *ClientEphemeralfiles) UploadEndpoint() string {
	return fmt.Sprintf("%s/%s/upload/clear", c.endpoint, apiVersion)
}

// Upload uploads a file to the ephemeralfiles service.
func (c *ClientEphemeralfiles) Upload(fileToUpload string) error {
	stat, err := c.validateAndGetFileInfo(fileToUpload)
	if err != nil {
		return err
	}

	c.InitProgressBar("uploading file...", stat.Size())
	defer c.CloseProgressBar()

	// Create multipart form and upload
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go c.createMultipartForm(writer, pw, fileToUpload)

	return c.sendUploadRequest(pr, writer)
}

// validateAndGetFileInfo validates file existence and returns file info.
func (c *ClientEphemeralfiles) validateAndGetFileInfo(fileToUpload string) (os.FileInfo, error) {
	stat, err := os.Stat(fileToUpload)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrFileNotFound, fileToUpload)
		}
		return nil, fmt.Errorf("error getting file info: %w", err)
	}
	return stat, nil
}

// createMultipartForm creates and populates the multipart form.
func (c *ClientEphemeralfiles) createMultipartForm(writer *multipart.Writer, pw *io.PipeWriter, fileToUpload string) {
	defer func() {
		_ = pw.Close()
	}()

	part, err := writer.CreateFormFile("uploadfile", fileToUpload)
	if err != nil {
		pw.CloseWithError(err)
		return
	}

	// #nosec G304 -- fileToUpload is provided by user for file upload
	f, err := os.Open(fileToUpload)
	if err != nil {
		pw.CloseWithError(err)
		return
	}
	defer func() {
		_ = f.Close()
	}()

	if _, err := io.Copy(io.MultiWriter(part, c.bar), f); err != nil {
		pw.CloseWithError(err)
		return
	}

	_ = writer.Close()
}

// sendUploadRequest creates and sends the upload HTTP request.
func (c *ClientEphemeralfiles) sendUploadRequest(pr *io.PipeReader, writer *multipart.Writer) error {
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.UploadEndpoint(), pr)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+c.token)
	req.Header.Add("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
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
