package parser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// PackageJSON represents the structure of package.json
type PackageJSON struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

// PackageLockJSON represents npm's package-lock.json structure
type PackageLockJSON struct {
	Packages map[string]PackageLockEntry `json:"packages"`
}

type PackageLockEntry struct {
	Version string `json:"version"`
}

// PnpmLockYAML represents pnpm's pnpm-lock.yaml structure (simplified)
type PnpmLockEntry struct {
	Name    string
	Version string
}

// ParseNodeDeps parses dependencies from a Node.js project
// It attempts to parse lockfiles first (for exact versions) before falling back to package.json
func ParseNodeDeps(projectRoot string) ([]Dependency, error) {
	// Try parsing lockfiles first for exact versions
	deps, err := parseLockfile(projectRoot)
	if err == nil && len(deps) > 0 {
		return deps, nil
	}

	// Fall back to package.json
	return parsePackageJSON(projectRoot)
}

// parseLockfile attempts to parse various lockfile formats
func parseLockfile(projectRoot string) ([]Dependency, error) {
	// Try package-lock.json (npm)
	if deps, err := parseNpmLock(projectRoot); err == nil {
		return deps, nil
	}

	// Try yarn.lock
	if deps, err := parseYarnLock(projectRoot); err == nil {
		return deps, nil
	}

	// Try pnpm-lock.yaml
	if deps, err := parsePnpmLock(projectRoot); err == nil {
		return deps, nil
	}

	return nil, fmt.Errorf("no lockfile found")
}

// parseNpmLock parses npm's package-lock.json
func parseNpmLock(projectRoot string) ([]Dependency, error) {
	lockPath := filepath.Join(projectRoot, "package-lock.json")
	
	data, err := os.ReadFile(lockPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read package-lock.json: %w", err)
	}

	var lock PackageLockJSON
	if err := json.Unmarshal(data, &lock); err != nil {
		return nil, fmt.Errorf("failed to parse package-lock.json: %w", err)
	}

	deps := []Dependency{}
	seen := make(map[string]bool)

	for pkgPath, entry := range lock.Packages {
		// Skip the root package (empty string key)
		if pkgPath == "" {
			continue
		}

		// Extract package name from path (e.g., "node_modules/express" -> "express")
		name := strings.TrimPrefix(pkgPath, "node_modules/")
		
		// Handle scoped packages and nested dependencies
		// We only want direct dependencies at the first level
		if strings.Count(name, "/") > 1 || (strings.HasPrefix(name, "@") && strings.Count(name, "/") > 2) {
			continue
		}

		if !seen[name] && entry.Version != "" {
			deps = append(deps, Dependency{
				Name:    name,
				Version: entry.Version,
				Type:    "npm",
			})
			seen[name] = true
		}
	}

	if len(deps) == 0 {
		return nil, fmt.Errorf("no dependencies found in package-lock.json")
	}

	return deps, nil
}

// parseYarnLock parses yarn.lock
func parseYarnLock(projectRoot string) ([]Dependency, error) {
	lockPath := filepath.Join(projectRoot, "yarn.lock")
	
	file, err := os.Open(lockPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read yarn.lock: %w", err)
	}
	defer file.Close()

	deps := []Dependency{}
	seen := make(map[string]bool)
	
	scanner := bufio.NewScanner(file)
	var currentPkg string
	var currentVersion string
	
	// Regex to match package declaration including scoped packages
	// Matches: "package@version:", package@version:, "@scope/package@version:"
	pkgRegex := regexp.MustCompile(`^"?(@?[^"@]+(?:\/[^"@]+)?)(?:@[^"]*)"?:?\s*$`)
	versionRegex := regexp.MustCompile(`^\s+version\s+"?([^"]+)"?\s*$`)

	for scanner.Scan() {
		line := scanner.Text()
		
		// Skip comments and empty lines
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}
		
		// Match package name
		if matches := pkgRegex.FindStringSubmatch(line); matches != nil {
			currentPkg = matches[1]
			currentVersion = ""
		} else if matches := versionRegex.FindStringSubmatch(line); matches != nil {
			// Match version
			currentVersion = matches[1]
			
			// Add dependency when we have both package and version
			if currentPkg != "" && currentVersion != "" && !seen[currentPkg] {
				deps = append(deps, Dependency{
					Name:    currentPkg,
					Version: currentVersion,
					Type:    "npm",
				})
				seen[currentPkg] = true
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading yarn.lock: %w", err)
	}

	if len(deps) == 0 {
		return nil, fmt.Errorf("no dependencies found in yarn.lock")
	}

	return deps, nil
}

// parsePnpmLock parses pnpm-lock.yaml
func parsePnpmLock(projectRoot string) ([]Dependency, error) {
	lockPath := filepath.Join(projectRoot, "pnpm-lock.yaml")
	
	file, err := os.Open(lockPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read pnpm-lock.yaml: %w", err)
	}
	defer file.Close()

	deps := []Dependency{}
	seen := make(map[string]bool)
	
	scanner := bufio.NewScanner(file)
	inPackages := false
	
	// Regex to match dependency entries like: /axios/1.4.0:
	// or /@types/node/20.0.0:
	pkgPathRegex := regexp.MustCompile(`^\s+/(@?[^/]+(?:/[^/]+)?)/([^:]+):`)

	for scanner.Scan() {
		line := scanner.Text()
		
		// Detect packages section
		if strings.HasPrefix(line, "packages:") {
			inPackages = true
			continue
		}
		
		// Exit sections when we hit another top-level key
		if len(line) > 0 && line[0] != ' ' && line[0] != '\t' && !strings.HasPrefix(line, "packages:") {
			inPackages = false
		}
		
		// Parse package paths like: /axios/1.4.0: or /@types/node/20.0.0:
		if inPackages {
			if matches := pkgPathRegex.FindStringSubmatch(line); matches != nil {
				name := matches[1]
				version := matches[2]
				
				if !seen[name] {
					deps = append(deps, Dependency{
						Name:    name,
						Version: version,
						Type:    "npm",
					})
					seen[name] = true
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading pnpm-lock.yaml: %w", err)
	}

	if len(deps) == 0 {
		return nil, fmt.Errorf("no dependencies found in pnpm-lock.yaml")
	}

	return deps, nil
}

// parsePackageJSON parses package.json for dependencies (fallback method)
func parsePackageJSON(projectRoot string) ([]Dependency, error) {
	packageJSONPath := filepath.Join(projectRoot, "package.json")
	
	data, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read package.json: %w", err)
	}

	var pkg PackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("failed to parse package.json: %w", err)
	}

	deps := []Dependency{}

	// Add regular dependencies
	for name, version := range pkg.Dependencies {
		deps = append(deps, Dependency{
			Name:    name,
			Version: cleanVersion(version),
			Type:    "npm",
		})
	}

	// Add dev dependencies
	for name, version := range pkg.DevDependencies {
		deps = append(deps, Dependency{
			Name:    name,
			Version: cleanVersion(version),
			Type:    "npm",
		})
	}

	return deps, nil
}

// cleanVersion removes version prefixes like ^, ~, >=, etc.
func cleanVersion(version string) string {
	// Remove common version prefixes
	prefixes := []string{"^", "~", ">=", "<=", ">", "<", "="}
	for _, prefix := range prefixes {
		if len(version) > len(prefix) && version[:len(prefix)] == prefix {
			version = version[len(prefix):]
		}
	}
	return version
}
