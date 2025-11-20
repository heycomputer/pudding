package docs

import (
	"fmt"
	"os/exec"

	"github.com/heycomputer/pudding/internal/web_docs"
	"github.com/heycomputer/pudding/internal/parser"
)
// BrowserOpener is a function type for opening URLs in a browser
type BrowserOpener func(url string) error

// CommandRunner is a function type for running external commands
type CommandRunner func(name string, args ...string) error

// Default implementations
var (
	defaultBrowserOpener BrowserOpener = openBrowser
	defaultCommandRunner CommandRunner = runCommand
)

// FetchAndOpen fetches documentation for a dependency and opens it in the browser
func FetchAndOpen(dep *parser.Dependency, projectType parser.ProjectType) error {
	return fetchAndOpenWithFuncs(dep, projectType, defaultCommandRunner, defaultBrowserOpener)
}

// fetchAndOpenWithFuncs allows dependency injection for testing
func fetchAndOpenWithFuncs(dep *parser.Dependency, projectType parser.ProjectType, cmdRunner CommandRunner, browserOpener BrowserOpener) error {
	switch projectType {
	case parser.ProjectTypeElixir:
		return fetchElixirDocsWithRunner(dep, cmdRunner)
	case parser.ProjectTypeJavaScript:
		return fetchNodeDocsWithOpener(dep, browserOpener)
	case parser.ProjectTypeRuby:
		return fetchRubyDocsWithOpener(dep, browserOpener)
	default:
		return fmt.Errorf("unsupported project type: %s", projectType)
	}
}

func fetchElixirDocsWithRunner(dep *parser.Dependency, cmdRunner CommandRunner) error {
	// Use mix hex.docs offline to fetch and open docs
	if dep.Version != "" {
		return cmdRunner("mix", "hex.docs", "offline", dep.Name, dep.Version)
	}
	return cmdRunner("mix", "hex.docs", "offline", dep.Name)
}

func fetchNodeDocsWithOpener(dep *parser.Dependency, browserOpener BrowserOpener) error {
	// For npm packages, we'll open the npm package page
	url := fmt.Sprintf("https://www.npmjs.com/package/%s/v/%s", dep.Name, dep.Version)
	return browserOpener(url)
}

func fetchRubyDocsWithOpener(dep *parser.Dependency, browserOpener BrowserOpener) error {
	// Create downloader instance
	downloader := web_docs.NewDownloader("")

	// Use RubyGems API to get the best documentation URL
	client := NewRubyGemsAPIClient()
	url, err := client.GetDocumentationURL(dep.Name, dep.Version)
	if err != nil {
		// Fallback to rubygems.org page if API call fails
		url = fmt.Sprintf("https://rubygems.org/gems/%s/versions/%s", dep.Name, dep.Version)
	}

	// Download and cache the documentation site
	result, err := downloader.Download(url, "ruby", dep.Name, dep.Version)
	if err != nil {
		// If download fails, fall back to opening in browser
		return browserOpener(url)
	}

	// Open the cached documentation in the browser
	return browserOpener("file://" + result.IndexPath)
}

// runCommand executes an external command
func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run %s: %w (output: %s)", name, err, string(output))
	}
	return nil
}

func openBrowser(url string) error {
	// Try to open URL in default browser
	// macOS: open, Linux: xdg-open, Windows: start
	cmd := exec.Command("open", url)
	err := cmd.Start()
	if err != nil {
		// Try xdg-open for Linux
		cmd = exec.Command("xdg-open", url)
		err = cmd.Start()
	}
	return err
}
