package docs

import (
	"fmt"
	"strings"
	"testing"

	"github.com/heycomputer/pudd/internal/parser"
)

// Mock implementations for testing
var (
	lastCommandRun  string
	lastBrowserURL  string
	commandShouldFail bool
	browserShouldFail bool
)

func mockCommandRunner(name string, args ...string) error {
	lastCommandRun = fmt.Sprintf("%s %s", name, strings.Join(args, " "))
	if commandShouldFail {
		return fmt.Errorf("mock command error")
	}
	return nil
}

func mockBrowserOpener(url string) error {
	lastBrowserURL = url
	if browserShouldFail {
		return fmt.Errorf("mock browser error")
	}
	return nil
}

func resetMocks() {
	lastCommandRun = ""
	lastBrowserURL = ""
	commandShouldFail = false
	browserShouldFail = false
}

func TestFetchAndOpen_UnsupportedProjectType(t *testing.T) {
	resetMocks()
	dep := &parser.Dependency{
		Name:    "test",
		Version: "1.0.0",
		Type:    "unknown",
	}

	err := fetchAndOpenWithFuncs(dep, parser.ProjectTypeUnknown, mockCommandRunner, mockBrowserOpener)
	if err == nil {
		t.Errorf("Expected error for unsupported project type, got nil")
	}
	
	if !strings.Contains(err.Error(), "unsupported project type") {
		t.Errorf("Expected 'unsupported project type' error, got: %v", err)
	}
}

func TestFetchElixirDocs(t *testing.T) {
	tests := []struct {
		name            string
		dep             *parser.Dependency
		expectedCommand string
	}{
		{
			name: "With version",
			dep: &parser.Dependency{
				Name:    "phoenix",
				Version: "1.7.0",
				Type:    "elixir",
			},
			expectedCommand: "mix hex.docs offline phoenix 1.7.0",
		},
		{
			name: "Without version",
			dep: &parser.Dependency{
				Name:    "ecto",
				Version: "",
				Type:    "elixir",
			},
			expectedCommand: "mix hex.docs offline ecto",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetMocks()
			
			err := fetchAndOpenWithFuncs(tt.dep, parser.ProjectTypeElixir, mockCommandRunner, mockBrowserOpener)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			if lastCommandRun != tt.expectedCommand {
				t.Errorf("Expected command '%s', got '%s'", tt.expectedCommand, lastCommandRun)
			}
			
			if lastBrowserURL != "" {
				t.Errorf("Expected no browser call, but got URL: %s", lastBrowserURL)
			}
		})
	}
}

func TestFetchElixirDocs_CommandFailure(t *testing.T) {
	resetMocks()
	commandShouldFail = true
	
	dep := &parser.Dependency{
		Name:    "phoenix",
		Version: "1.7.0",
		Type:    "elixir",
	}
	
	err := fetchAndOpenWithFuncs(dep, parser.ProjectTypeElixir, mockCommandRunner, mockBrowserOpener)
	if err == nil {
		t.Errorf("Expected error when command fails, got nil")
	}
}

func TestFetchNodeDocs(t *testing.T) {
	resetMocks()
	
	dep := &parser.Dependency{
		Name:    "express",
		Version: "4.18.2",
		Type:    "npm",
	}
	
	err := fetchAndOpenWithFuncs(dep, parser.ProjectTypeJavaScript, mockCommandRunner, mockBrowserOpener)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	expectedURL := "https://www.npmjs.com/package/express/v/4.18.2"
	if lastBrowserURL != expectedURL {
		t.Errorf("Expected URL '%s', got '%s'", expectedURL, lastBrowserURL)
	}
	
	if lastCommandRun != "" {
		t.Errorf("Expected no command run, but got: %s", lastCommandRun)
	}
}

func TestFetchNodeDocs_BrowserFailure(t *testing.T) {
	resetMocks()
	browserShouldFail = true
	
	dep := &parser.Dependency{
		Name:    "express",
		Version: "4.18.2",
		Type:    "npm",
	}
	
	err := fetchAndOpenWithFuncs(dep, parser.ProjectTypeJavaScript, mockCommandRunner, mockBrowserOpener)
	if err == nil {
		t.Errorf("Expected error when browser fails, got nil")
	}
}

func TestFetchRubyDocs(t *testing.T) {
	resetMocks()
	
	dep := &parser.Dependency{
		Name:    "rails",
		Version: "7.0.0",
		Type:    "gem",
	}
	
	err := fetchAndOpenWithFuncs(dep, parser.ProjectTypeRuby, mockCommandRunner, mockBrowserOpener)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	// The new implementation uses RubyGems API which will try to fetch the gem info
	// In testing without a mock server, it will fail and fallback to rubydoc.info
	// We just verify that a URL was opened
	if lastBrowserURL == "" {
		t.Errorf("Expected a browser URL to be opened, got empty string")
	}
	
	if lastCommandRun != "" {
		t.Errorf("Expected no command run, but got: %s", lastCommandRun)
	}
}

func TestFetchRubyDocs_BrowserFailure(t *testing.T) {
	resetMocks()
	browserShouldFail = true
	
	dep := &parser.Dependency{
		Name:    "rails",
		Version: "7.0.0",
		Type:    "gem",
	}
	
	err := fetchAndOpenWithFuncs(dep, parser.ProjectTypeRuby, mockCommandRunner, mockBrowserOpener)
	if err == nil {
		t.Errorf("Expected error when browser fails, got nil")
	}
}

func TestAllProjectTypes(t *testing.T) {
	tests := []struct {
		name        string
		projectType parser.ProjectType
		dep         *parser.Dependency
		expectCmd   bool
		expectURL   bool
	}{
		{
			name:        "Elixir",
			projectType: parser.ProjectTypeElixir,
			dep: &parser.Dependency{
				Name:    "phoenix",
				Version: "1.7.0",
				Type:    "elixir",
			},
			expectCmd: true,
			expectURL: false,
		},
		{
			name:        "JavaScript",
			projectType: parser.ProjectTypeJavaScript,
			dep: &parser.Dependency{
				Name:    "lodash",
				Version: "4.17.21",
				Type:    "npm",
			},
			expectCmd: false,
			expectURL: true,
		},
		{
			name:        "Ruby",
			projectType: parser.ProjectTypeRuby,
			dep: &parser.Dependency{
				Name:    "sinatra",
				Version: "3.0.0",
				Type:    "gem",
			},
			expectCmd: false,
			expectURL: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetMocks()
			
			err := fetchAndOpenWithFuncs(tt.dep, tt.projectType, mockCommandRunner, mockBrowserOpener)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			if tt.expectCmd && lastCommandRun == "" {
				t.Errorf("Expected command to be run, but it wasn't")
			}
			if !tt.expectCmd && lastCommandRun != "" {
				t.Errorf("Expected no command, but got: %s", lastCommandRun)
			}
			
			if tt.expectURL && lastBrowserURL == "" {
				t.Errorf("Expected browser URL, but it wasn't set")
			}
			if !tt.expectURL && lastBrowserURL != "" {
				t.Errorf("Expected no browser URL, but got: %s", lastBrowserURL)
			}
		})
	}
}
