package parser

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// ParseElixirDeps parses dependencies from a Mix project
func ParseElixirDeps(projectRoot string) ([]Dependency, error) {
	// Use mix deps command to get dependencies
	cmd := exec.Command("mix", "deps", "--all")
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to run mix deps: %w (output: %s)", err, string(output))
	}

	// Also get Elixir version
	elixirVersion, err := getElixirVersion(projectRoot)
	if err != nil {
		// Non-fatal, just skip Elixir entry
		elixirVersion = ""
	}

	deps := []Dependency{}

	// Add Elixir itself as a dependency if we got the version
	if elixirVersion != "" {
		deps = append(deps, Dependency{
			Name:    "elixir",
			Version: elixirVersion,
			Type:    "elixir",
		})
	}

	// Parse mix deps output
	// Format is typically:
	// * dep_name (hex package) (mix)
	//   locked at 1.2.3 (dep_name) hash
	lines := strings.Split(string(output), "\n")
	depNameRegex := regexp.MustCompile(`^\* (\S+)`)
	versionRegex := regexp.MustCompile(`locked at (\S+)`)

	var currentDep string
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Check if this is a dependency name line
		if matches := depNameRegex.FindStringSubmatch(line); len(matches) > 1 {
			currentDep = matches[1]
		} else if currentDep != "" && strings.Contains(line, "locked at") {
			// This is the version line
			if matches := versionRegex.FindStringSubmatch(line); len(matches) > 1 {
				deps = append(deps, Dependency{
					Name:    currentDep,
					Version: matches[1],
					Type:    "elixir",
				})
			}
			currentDep = ""
		}
	}

	return deps, nil
}

func getElixirVersion(projectRoot string) (string, error) {
	cmd := exec.Command("mix", "hex.info")
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	// Parse output for "Elixir: 1.x.x"
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "Elixir:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1], nil
			}
		}
	}

	return "", fmt.Errorf("elixir version not found")
}
