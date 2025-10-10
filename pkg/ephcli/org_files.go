// Package ephcli provides organization file operation methods for the ephemeralfiles API.
package ephcli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/ephemeralfiles/eph/pkg/dto"
)

// UploadOrganizationFile uploads a file to an organization.
func (c *ClientEphemeralfiles) UploadOrganizationFile(
	orgID string,
	filepath string,
	tags []string,
) (*dto.OrganizationFile, error) {
	stat, err := c.validateAndGetFileInfo(filepath)
	if err != nil {
		return nil, err
	}

	c.InitProgressBar("uploading file...", stat.Size())
	defer c.CloseProgressBar()

	// Create multipart form and upload
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go c.createOrgMultipartForm(writer, pw, filepath, tags)

	file, err := c.sendOrgUploadRequest(orgID, pr, writer)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// createOrgMultipartForm creates and populates the multipart form for organization upload.
func (c *ClientEphemeralfiles) createOrgMultipartForm(
	writer *multipart.Writer,
	pw *io.PipeWriter,
	filepath string,
	tags []string,
) {
	defer func() {
		_ = pw.Close()
	}()

	// Add file
	part, err := writer.CreateFormFile("file", filepath)
	if err != nil {
		c.log.Debug("createOrgMultipartForm: CreateFormFile failed", slog.String("error", err.Error()))
		pw.CloseWithError(err)
		return
	}

	// #nosec G304 -- filepath is provided by user for file upload
	f, err := os.Open(filepath)
	if err != nil {
		c.log.Debug("createOrgMultipartForm: Open file failed", slog.String("error", err.Error()))
		pw.CloseWithError(err)
		return
	}
	defer func() {
		_ = f.Close()
	}()

	stat, _ := f.Stat()
	c.log.Debug("createOrgMultipartForm: About to copy file",
		slog.String("filepath", filepath),
		slog.Int64("size", stat.Size()))

	bytesWritten, err := io.Copy(io.MultiWriter(part, c.bar), f)
	if err != nil {
		c.log.Debug("createOrgMultipartForm: Copy failed", slog.String("error", err.Error()))
		pw.CloseWithError(err)
		return
	}

	c.log.Debug("createOrgMultipartForm: File copied",
		slog.Int64("bytesWritten", bytesWritten))

	// Add tags if provided
	if len(tags) > 0 {
		if err := writer.WriteField("tags", strings.Join(tags, ",")); err != nil {
			c.log.Debug("createOrgMultipartForm: WriteField tags failed", slog.String("error", err.Error()))
			pw.CloseWithError(err)
			return
		}
	}

	if err := writer.Close(); err != nil {
		c.log.Debug("createOrgMultipartForm: writer.Close failed", slog.String("error", err.Error()))
		pw.CloseWithError(err)
		return
	}

	c.log.Debug("createOrgMultipartForm: Multipart form completed successfully")
}

// sendOrgUploadRequest creates and sends the organization upload HTTP request.
func (c *ClientEphemeralfiles) sendOrgUploadRequest(
	orgID string,
	pr *io.PipeReader,
	writer *multipart.Writer,
) (*dto.OrganizationFile, error) {
	uploadURL := fmt.Sprintf("%s/%s/organizations/%s/files/upload", c.endpoint, apiVersion, orgID)

	ctx, cancel := context.WithTimeout(context.Background(), ChunkUploadTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, pr)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+c.token)
	req.Header.Add("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.log.Debug("Warning: failed to close response body", slog.String("error", closeErr.Error()))
		}
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, parseError(resp)
	}

	var file dto.OrganizationFile
	if err := json.NewDecoder(resp.Body).Decode(&file); err != nil {
		return nil, fmt.Errorf("failed to decode upload response: %w", err)
	}

	return &file, nil
}

// ListOrganizationFiles lists all files in an organization.
func (c *ClientEphemeralfiles) ListOrganizationFiles(
	orgID string,
	limit int,
	offset int,
) ([]dto.OrganizationFile, error) {
	urlStr := fmt.Sprintf("%s/%s/organizations/%s/files?limit=%d&offset=%d",
		c.endpoint, apiVersion, orgID, limit, offset)

	req, cancel, err := c.createRequestWithTimeout(http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}
	defer cancel()

	resp, err := c.doRequestWithAuth(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var files []dto.OrganizationFile
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, fmt.Errorf("failed to decode files response: %w", err)
	}

	return files, nil
}

// GetOrganizationFilesByTags gets files filtered by tags.
func (c *ClientEphemeralfiles) GetOrganizationFilesByTags(
	orgID string,
	tags []string,
	limit int,
	offset int,
) ([]dto.OrganizationFile, error) {
	tagsParam := url.QueryEscape(strings.Join(tags, ","))
	urlStr := fmt.Sprintf("%s/%s/organizations/%s/files/tags?tags=%s&limit=%d&offset=%d",
		c.endpoint, apiVersion, orgID, tagsParam, limit, offset)

	req, cancel, err := c.createRequestWithTimeout(http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}
	defer cancel()

	resp, err := c.doRequestWithAuth(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var files []dto.OrganizationFile
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, fmt.Errorf("failed to decode files response: %w", err)
	}

	return files, nil
}

// ListRecentOrganizationFiles lists recent files in an organization.
func (c *ClientEphemeralfiles) ListRecentOrganizationFiles(
	orgID string,
	limit int,
) ([]dto.OrganizationFile, error) {
	urlStr := fmt.Sprintf("%s/%s/organizations/%s/files/recent?limit=%d",
		c.endpoint, apiVersion, orgID, limit)

	req, cancel, err := c.createRequestWithTimeout(http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}
	defer cancel()

	resp, err := c.doRequestWithAuth(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var files []dto.OrganizationFile
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, fmt.Errorf("failed to decode files response: %w", err)
	}

	return files, nil
}

// ListExpiredOrganizationFiles lists expired files in an organization.
func (c *ClientEphemeralfiles) ListExpiredOrganizationFiles(
	orgID string,
	limit int,
) ([]dto.OrganizationFile, error) {
	urlStr := fmt.Sprintf("%s/%s/organizations/%s/files/expired?limit=%d",
		c.endpoint, apiVersion, orgID, limit)

	req, cancel, err := c.createRequestWithTimeout(http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}
	defer cancel()

	resp, err := c.doRequestWithAuth(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var files []dto.OrganizationFile
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, fmt.Errorf("failed to decode files response: %w", err)
	}

	return files, nil
}

// GetPopularTags retrieves popular tags for an organization.
func (c *ClientEphemeralfiles) GetPopularTags(
	orgID string,
	limit int,
) ([]dto.TagCount, error) {
	urlStr := fmt.Sprintf("%s/%s/organizations/%s/tags?limit=%d",
		c.endpoint, apiVersion, orgID, limit)

	req, cancel, err := c.createRequestWithTimeout(http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}
	defer cancel()

	resp, err := c.doRequestWithAuth(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var tags []dto.TagCount
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, fmt.Errorf("failed to decode tags response: %w", err)
	}

	return tags, nil
}

// DeleteOrganizationFile deletes a file from an organization.
func (c *ClientEphemeralfiles) DeleteOrganizationFile(fileID string) error {
	urlStr := fmt.Sprintf("%s/%s/files/%s", c.endpoint, apiVersion, fileID)

	req, cancel, err := c.createRequestWithTimeout(http.MethodDelete, urlStr, nil)
	if err != nil {
		return err
	}
	defer cancel()

	resp, err := c.doRequestWithAuth(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	return nil
}

// UpdateFileTags updates tags for a file.
func (c *ClientEphemeralfiles) UpdateFileTags(
	fileID string,
	tags []string,
) (*dto.OrganizationFile, error) {
	urlStr := fmt.Sprintf("%s/%s/files/%s/tags", c.endpoint, apiVersion, fileID)

	tagsJSON, err := json.Marshal(map[string][]string{"tags": tags})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tags: %w", err)
	}

	req, cancel, err := c.createRequestWithTimeout(http.MethodPut, urlStr, strings.NewReader(string(tagsJSON)))
	if err != nil {
		return nil, err
	}
	defer cancel()

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.doRequestWithAuth(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var file dto.OrganizationFile
	if err := json.NewDecoder(resp.Body).Decode(&file); err != nil {
		return nil, fmt.Errorf("failed to decode file response: %w", err)
	}

	return &file, nil
}

// DownloadOrganizationFile downloads a file from an organization.
func (c *ClientEphemeralfiles) DownloadOrganizationFile(fileID string, outputFile string) error {
	// Organization files use the authenticated files endpoint
	urlStr := fmt.Sprintf("%s/%s/files/%s/download", c.endpoint, apiVersion, fileID)

	req, cancel, err := c.createRequestWithTimeout(http.MethodGet, urlStr, nil)
	if err != nil {
		return err
	}
	defer cancel()

	resp, err := c.doRequestWithAuth(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return parseError(resp)
	}

	// Get filename from Content-Disposition header or use output file
	filename := c.getFileName(resp, outputFile)

	// Create the output file
	// #nosec G304 -- filename is derived from server response headers for downloads
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	// Set up progress bar
	totalSize := resp.ContentLength
	c.InitProgressBar("downloading file...", totalSize)
	defer c.CloseProgressBar()

	// Copy the file content
	if !c.noProgressBar {
		_, err = io.Copy(io.MultiWriter(f, c.bar), resp.Body)
	} else {
		_, err = io.Copy(f, resp.Body)
	}
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}
