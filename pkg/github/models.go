package github

// ResponseLatestRelease represents the response from GitHub's latest release API.
type ResponseLatestRelease struct {
	TagName string `json:"tag_name"`
}
