package selector

import (
	"testing"

	"github.com/heycomputer/pudd/internal/parser"
)

func TestFilterDependencies(t *testing.T) {
	deps := []parser.Dependency{
		{Name: "express", Version: "4.18.2", Type: "npm"},
		{Name: "lodash", Version: "4.17.21", Type: "npm"},
		{Name: "axios", Version: "1.4.0", Type: "npm"},
		{Name: "jest", Version: "29.5.0", Type: "npm"},
		{Name: "eslint", Version: "8.0.0", Type: "npm"},
	}

	tests := []struct {
		name     string
		query    string
		expected int
	}{
		{"Empty query returns all", "", 5},
		{"Exact match", "express", 1},
		{"Partial match", "es", 3}, // express, jest, eslint
		{"Case insensitive", "EXPRESS", 1},
		{"No match", "nonexistent", 0},
		{"Multiple matches", "e", 3}, // express, jest, eslint
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterDependencies(deps, tt.query)
			if len(result) != tt.expected {
				t.Errorf("FilterDependencies(%q) returned %d results, expected %d",
					tt.query, len(result), tt.expected)
			}
		})
	}
}

func TestFilterDependencies_EmptyList(t *testing.T) {
	deps := []parser.Dependency{}
	result := FilterDependencies(deps, "anything")
	
	if len(result) != 0 {
		t.Errorf("Expected 0 results for empty dependency list, got %d", len(result))
	}
}

func TestFilterDependencies_PreservesOrder(t *testing.T) {
	deps := []parser.Dependency{
		{Name: "zebra", Version: "1.0.0", Type: "npm"},
		{Name: "alpha", Version: "2.0.0", Type: "npm"},
		{Name: "beta", Version: "3.0.0", Type: "npm"},
	}

	result := FilterDependencies(deps, "")
	
	// Should preserve original order
	if len(result) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(result))
	}
	
	if result[0].Name != "zebra" || result[1].Name != "alpha" || result[2].Name != "beta" {
		t.Errorf("FilterDependencies did not preserve order")
	}
}

func TestFilterDependencies_CaseInsensitive(t *testing.T) {
	deps := []parser.Dependency{
		{Name: "Express", Version: "4.18.2", Type: "npm"},
		{Name: "LODASH", Version: "4.17.21", Type: "npm"},
		{Name: "AxIoS", Version: "1.4.0", Type: "npm"},
	}

	tests := []struct {
		query    string
		expected int
	}{
		{"express", 1},
		{"EXPRESS", 1},
		{"ExPrEsS", 1},
		{"lodash", 1},
		{"AXIOS", 1},
		{"axios", 1},
	}

	for _, tt := range tests {
		result := FilterDependencies(deps, tt.query)
		if len(result) != tt.expected {
			t.Errorf("FilterDependencies(%q) returned %d results, expected %d",
				tt.query, len(result), tt.expected)
		}
	}
}
