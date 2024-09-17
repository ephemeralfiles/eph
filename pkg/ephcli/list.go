package ephcli

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/pterm/pterm"
	"gopkg.in/yaml.v2"
)

// FilesEndpoint returns the endpoint for the files
func (c *ClientEphemeralfiles) FilesEndpoint() string {
	return fmt.Sprintf("%s/%s/files", c.endpoint, apiVersion)
}

// Fetch retrieves the list of files from the server
func (c *ClientEphemeralfiles) Fetch() (FileList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultAPIRequestTimeout)
	defer cancel()

	url := c.FilesEndpoint()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, parseError(resp)
	}

	var fl FileList
	err = json.NewDecoder(resp.Body).Decode(&fl)
	if err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	if len(fl) == 0 {
		return nil, nil
	}
	return fl, nil
}

// Print prints the list of files as a table
func (fl *FileList) Print() error {
	tData := pterm.TableData{
		{"ID", "Filename", "Size", "Expiration date"},
	}
	for _, file := range *fl {
		tData = append(tData, []string{file.Idfile, file.FileName,
			strconv.FormatInt(file.Size, 10), file.ExpirationDate.Format("2006-01-02 15:04:05")})
	}
	// Create a table with a header and the defined data, then render it
	err := pterm.DefaultTable.WithHasHeader().WithData(tData).Render()
	if err != nil {
		return fmt.Errorf("error rendering table: %w", err)
	}
	return nil
}

// PrintCSV prints the list of files as a CSV
func (fl *FileList) PrintCSV() error {
	csvwriter := csv.NewWriter(os.Stdout)
	err := csvwriter.Write([]string{"ID", "Filename", "Size", "Expiration date"})
	if err != nil {
		return fmt.Errorf("error writing CSV header: %w", err)
	}

	for _, file := range *fl {
		err = csvwriter.Write([]string{file.Idfile, file.FileName,
			strconv.FormatInt(file.Size, 10), file.ExpirationDate.Format("2006-01-02 15:04:05")})
		if err != nil {
			return fmt.Errorf("error writing CSV row: %w", err)
		}
	}

	csvwriter.Flush()
	return nil
}

// PrintJSON prints the list of files as JSON
func (fl *FileList) PrintJSON() error {
	err := json.NewEncoder(os.Stdout).Encode(fl)
	if err != nil {
		return fmt.Errorf("error encoding JSON: %w", err)
	}
	return nil
}

// PrintYAML prints the list of files as YAML
func (fl *FileList) PrintYAML() error {
	yamlData, err := yaml.Marshal(fl)
	if err != nil {
		return fmt.Errorf("error marshalling YAML: %w", err)
	}
	fmt.Println(string(yamlData))
	return nil
}
