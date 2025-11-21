package docs

import (
	"fmt"
	"os/exec"
	"strings"
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

func fetchRubyDocsWithOpener(dep *parser.Dependency, browserOpener BrowserOpener) error {
	// Generate documentation using rdoc
	// rdoc GEM_NAME --rdoc --version GEM_VERSION
	cmdRunner := defaultCommandRunner
	if err := cmdRunner("rdoc", dep.Name, "--rdoc", "--version", dep.Version); err != nil {
		return fmt.Errorf("failed to generate rdoc for %s: %w", dep.Name, err)
	}

	// run command to get gem env home and assign to variable
	gemEnvOutput, err := exec.Command("sh", "-c", "gem env home").Output()
	if err != nil {
		return fmt.Errorf("failed to get gem env home: %w", err)
	}
	// convert gemEnvOutput to string and strip whitespace/newlines
	gemEnvHome := string(gemEnvOutput)
	gemEnvHome = strings.TrimSpace(gemEnvHome)

	// Get the path to the generated documentation
	// open $(gem env home)/doc/GEM_NAME-GEM_VERSION/rdoc/table_of_contents.html
	gemDocTocUrl := fmt.Sprintf("%s/doc/%s-%s/rdoc/table_of_contents.html", gemEnvHome, dep.Name, dep.Version)

	// Open the documentation in browser using shell expansion
	if err := browserOpener(gemDocTocUrl); err != nil {
		return fmt.Errorf("failed to open rdoc for %s: %w", dep.Name, err)
	}

	return nil
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
