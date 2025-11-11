package docs

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// RubyGemsAPIClient interacts with the RubyGems.org API V2
type RubyGemsAPIClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewRubyGemsAPIClient creates a new RubyGems API client
func NewRubyGemsAPIClient() *RubyGemsAPIClient {
	return &RubyGemsAPIClient{
		baseURL: "https://rubygems.org/api/v2",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GemVersion represents a specific version of a gem from the API
type GemVersion struct {
	Number           string `json:"number"`
	PreRelease       bool   `json:"prerelease"`
	Platform         string `json:"platform"`
	RubyVersion      string `json:"ruby_version"`
	RubyGemsVersion  string `json:"rubygems_version"`
	Summary          string `json:"summary"`
	Description      string `json:"description"`
	AuthorsArray     string `json:"authors"`
	License          string `json:"license"`
	Licenses         []string `json:"licenses"`
	Requirements     string `json:"requirements"`
	SHA              string `json:"sha"`
	Downloads        int    `json:"downloads_count"`
	Metadata         GemMetadata `json:"metadata"`
}

// GemMetadata contains metadata about a gem version
type GemMetadata struct {
	Homepage       string `json:"homepage_uri"`
	SourceCode     string `json:"source_code_uri"`
	Documentation  string `json:"documentation_uri"`
	ChangeLog      string `json:"changelog_uri"`
	BugTracker     string `json:"bug_tracker_uri"`
	MailingList    string `json:"mailing_list_uri"`
	Wiki           string `json:"wiki_uri"`
}

// GemInfo represents basic information about a gem
type GemInfo struct {
	Name             string `json:"name"`
	Downloads        int    `json:"downloads"`
	Version          string `json:"version"`
	VersionDownloads int    `json:"version_downloads"`
	Platform         string `json:"platform"`
	Authors          string `json:"authors"`
	Info             string `json:"info"`
	Licenses         []string `json:"licenses"`
	Metadata         GemMetadata `json:"metadata"`
	SHA              string `json:"sha"`
	ProjectURI       string `json:"project_uri"`
	GemURI           string `json:"gem_uri"`
	HomepageURI      string `json:"homepage_uri"`
	WikiURI          string `json:"wiki_uri"`
	DocumentationURI string `json:"documentation_uri"`
	MailingListURI   string `json:"mailing_list_uri"`
	SourceCodeURI    string `json:"source_code_uri"`
	BugTrackerURI    string `json:"bug_tracker_uri"`
	ChangelogURI     string `json:"changelog_uri"`
	Dependencies     struct {
		Development []interface{} `json:"development"`
		Runtime     []interface{} `json:"runtime"`
	} `json:"dependencies"`
}

// GetGemVersion fetches information about a specific version of a gem
func (c *RubyGemsAPIClient) GetGemVersion(name, version string) (*GemVersion, error) {
	url := fmt.Sprintf("%s/rubygems/%s/versions/%s.json", c.baseURL, name, version)
	
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch gem version: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d for gem %s version %s", resp.StatusCode, name, version)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var gemVersion GemVersion
	if err := json.Unmarshal(body, &gemVersion); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	return &gemVersion, nil
}

// GetGemInfo fetches general information about a gem (latest version)
func (c *RubyGemsAPIClient) GetGemInfo(name string) (*GemInfo, error) {
	url := fmt.Sprintf("%s/rubygems/%s.json", c.baseURL, name)
	
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch gem info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d for gem %s", resp.StatusCode, name)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var gemInfo GemInfo
	if err := json.Unmarshal(body, &gemInfo); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	return &gemInfo, nil
}

// GetDocumentationURL attempts to find the best documentation URL for a gem version
// It tries multiple sources in order of preference:
// 1. Version-specific metadata documentation_uri
// 2. Gem info documentation_uri (latest version)
// 3. Version-specific metadata homepage_uri
// 4. Gem info homepage_uri
// 5. Default to rubydoc.info (community documentation)
// 6. Fallback to rubygems.org page
func (c *RubyGemsAPIClient) GetDocumentationURL(name, version string) (string, error) {
	// Try to get version-specific information
	gemVersion, err := c.GetGemVersion(name, version)
	if err == nil {
		// Check version-specific metadata for documentation URI
		if gemVersion.Metadata.Documentation != "" {
			return gemVersion.Metadata.Documentation, nil
		}
		// Check version-specific metadata for homepage
		if gemVersion.Metadata.Homepage != "" {
			return gemVersion.Metadata.Homepage, nil
		}
	}

	// Try to get general gem information
	gemInfo, err := c.GetGemInfo(name)
	if err == nil {
		// Check metadata documentation URI
		if gemInfo.Metadata.Documentation != "" {
			return gemInfo.Metadata.Documentation, nil
		}
		// Check top-level documentation URI
		if gemInfo.DocumentationURI != "" {
			return gemInfo.DocumentationURI, nil
		}
		// Check metadata homepage
		if gemInfo.Metadata.Homepage != "" {
			return gemInfo.Metadata.Homepage, nil
		}
		// Check top-level homepage
		if gemInfo.HomepageURI != "" {
			return gemInfo.HomepageURI, nil
		}
	}

	// Fallback to rubydoc.info which provides community documentation
	// This often has better formatted docs than the gem homepage
	return fmt.Sprintf("https://www.rubydoc.info/gems/%s/%s", name, version), nil
}
