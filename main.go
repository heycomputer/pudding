package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/heycomputer/pudd/internal/docs"
	"github.com/heycomputer/pudd/internal/parser"
	"github.com/heycomputer/pudd/internal/selector"
)

func main() {
	// Parse command line flags
	var query string
	flag.StringVar(&query, "q", "", "Query/filter for dependency name")
	flag.Parse()

	// If there's a positional argument, use it as the query
	if flag.NArg() > 0 {
		query = flag.Arg(0)
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get current directory: %v\n", err)
		os.Exit(1)
	}

	// Parse project dependencies
	deps, projectType, err := parser.ParseProjectDependencies(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(deps) == 0 {
		fmt.Fprintf(os.Stderr, "No dependencies found in project\n")
		os.Exit(1)
	}

	// Sort dependencies by name for better UX
	sort.Slice(deps, func(i, j int) bool {
		return deps[i].Name < deps[j].Name
	})

	// Filter dependencies if query is provided
	filteredDeps := deps
	if query != "" {
		filteredDeps = selector.FilterDependencies(deps, query)
		if len(filteredDeps) == 0 {
			fmt.Fprintf(os.Stderr, "No dependencies matching '%s' found\n", query)
			os.Exit(1)
		}
	}

	// Let user select a dependency
	selectedDep, err := selector.SelectDependency(filteredDeps, query)
	if err != nil {
		// User cancelled or error occurred
		fmt.Fprintf(os.Stderr, "Selection cancelled or error: %v\n", err)
		os.Exit(1)
	}

	// Fetch and open documentation
	fmt.Printf("Opening documentation for %s %s...\n", selectedDep.Name, selectedDep.Version)
	if err := docs.FetchAndOpen(selectedDep, projectType); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to open documentation: %v\n", err)
		os.Exit(1)
	}
}
