package selector

import (
	"fmt"
	"strings"

	"github.com/heycomputer/pudd/internal/parser"
	fuzzyfinder "github.com/ktr0731/go-fuzzyfinder"
)

// SelectDependency allows the user to interactively select a dependency from a list
// using fuzzy finding. If query is provided, it will be used as initial filter.
func SelectDependency(deps []parser.Dependency, query string) (*parser.Dependency, error) {
	if len(deps) == 0 {
		return nil, fmt.Errorf("no dependencies found")
	}

	// If there's an exact match for the query, return it
	if query != "" {
		for i := range deps {
			if strings.EqualFold(deps[i].Name, query) {
				return &deps[i], nil
			}
		}
	}

	// Use fuzzy finder for interactive selection
	idx, err := fuzzyfinder.Find(
		deps,
		func(i int) string {
			return fmt.Sprintf("%s %s", deps[i].Name, deps[i].Version)
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			return fmt.Sprintf("Package: %s\nVersion: %s\nType: %s",
				deps[i].Name,
				deps[i].Version,
				deps[i].Type)
		}),
		fuzzyfinder.WithPromptString("view docs for> "),
	)

	if err != nil {
		return nil, err
	}

	return &deps[idx], nil
}

// FilterDependencies returns dependencies matching the query string (case-insensitive)
func FilterDependencies(deps []parser.Dependency, query string) []parser.Dependency {
	if query == "" {
		return deps
	}

	filtered := []parser.Dependency{}
	queryLower := strings.ToLower(query)

	for _, dep := range deps {
		if strings.Contains(strings.ToLower(dep.Name), queryLower) {
			filtered = append(filtered, dep)
		}
	}

	return filtered
}
