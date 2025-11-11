package parser

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// ParseRubyDeps parses dependencies from a Ruby project using bundler
func ParseRubyDeps(projectRoot string) ([]Dependency, error) {
	// First check if Gemfile.lock exists
	// If not, we might need to run bundle install first
	
	// Use bundle list to get dependencies with versions
	cmd := exec.Command("bundle", "list")
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to run bundle list: %w (output: %s)", err, string(output))
	}

	deps := []Dependency{}

	// Parse bundle list output
	// Format is typically:
	//   * gem_name (1.2.3)
	lines := strings.Split(string(output), "\n")
	gemRegex := regexp.MustCompile(`^\s*\*\s+(\S+)\s+\(([^)]+)\)`)

	for _, line := range lines {
		if matches := gemRegex.FindStringSubmatch(line); len(matches) > 2 {
			deps = append(deps, Dependency{
				Name:    matches[1],
				Version: matches[2],
				Type:    "gem",
			})
		}
	}

	// Also add Ruby version
	rubyVersion, err := getRubyVersion()
	if err == nil && rubyVersion != "" {
		// Prepend Ruby as first dependency
		deps = append([]Dependency{{
			Name:    "ruby",
			Version: rubyVersion,
			Type:    "gem",
		}}, deps...)
	}

	return deps, nil
}

func getRubyVersion() (string, error) {
	cmd := exec.Command("ruby", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	// Parse output like "ruby 3.2.0 (2022-12-25 revision a528908271) [x86_64-darwin22]"
	versionRegex := regexp.MustCompile(`ruby\s+(\d+\.\d+\.\d+)`)
	if matches := versionRegex.FindStringSubmatch(string(output)); len(matches) > 1 {
		return matches[1], nil
	}

	return "", fmt.Errorf("ruby version not found")
}
