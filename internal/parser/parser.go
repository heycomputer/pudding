package parser

import (
	"fmt"
	"os"
	"path/filepath"
)

// Dependency represents a single project dependency
type Dependency struct {
	Name    string
	Version string
	Type    string // "elixir", "gem"
}

// Parser interface for reading different dependency files
type Parser interface {
	Parse(projectRoot string) ([]Dependency, error)
	CanParse(projectRoot string) bool
}

// ProjectType represents the type of project
type ProjectType string

const (
	ProjectTypeElixir  ProjectType = "elixir"
	ProjectTypeRuby    ProjectType = "ruby"
	ProjectTypeUnknown ProjectType = "unknown"
)

// ParseProjectDependencies detects the project type and parses dependencies
func ParseProjectDependencies(dir string) ([]Dependency, ProjectType, error) {
	// Walk up the directory tree to find project root
	currentDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, ProjectTypeUnknown, err
	}

	for {
		// Check for mix.exs (Elixir)
		if fileExists(filepath.Join(currentDir, "mix.exs")) {
			deps, err := ParseElixirDeps(currentDir)
			return deps, ProjectTypeElixir, err
		}

		// Check for Gemfile (Ruby)
		if fileExists(filepath.Join(currentDir, "Gemfile")) {
			deps, err := ParseRubyDeps(currentDir)
			return deps, ProjectTypeRuby, err
		}

		// Move up one directory
		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			// Reached root directory
			break
		}
		currentDir = parent
	}

	return nil, ProjectTypeUnknown, fmt.Errorf("no supported project file found (mix.exs or Gemfile)")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
