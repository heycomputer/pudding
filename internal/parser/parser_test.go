package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileExists(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	
	// File doesn't exist yet
	if fileExists(tmpFile) {
		t.Errorf("fileExists returned true for non-existent file")
	}
	
	// Create the file
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// File should exist now
	if !fileExists(tmpFile) {
		t.Errorf("fileExists returned false for existing file")
	}
}

func TestParseProjectDependencies_NoProject(t *testing.T) {
	tmpDir := t.TempDir()
	
	_, projectType, err := ParseProjectDependencies(tmpDir)
	if err == nil {
		t.Errorf("Expected error for directory without project files, got nil")
	}
	if projectType != ProjectTypeUnknown {
		t.Errorf("Expected ProjectTypeUnknown, got %s", projectType)
	}
}

func TestParseProjectDependencies_NodeProject(t *testing.T) {
	tmpDir := t.TempDir()
	packageJSON := filepath.Join(tmpDir, "package.json")
	
	// Create a simple package.json
	content := `{
		"name": "test",
		"dependencies": {
			"express": "^4.18.2"
		}
	}`
	
	if err := os.WriteFile(packageJSON, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}
	
	deps, projectType, err := ParseProjectDependencies(tmpDir)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if projectType != ProjectTypeJavaScript {
		t.Errorf("Expected ProjectTypeJavaScript, got %s", projectType)
	}
	
	if len(deps) != 1 {
		t.Errorf("Expected 1 dependency, got %d", len(deps))
	}
	
	if len(deps) > 0 && deps[0].Name != "express" {
		t.Errorf("Expected dependency name 'express', got '%s'", deps[0].Name)
	}
}

func TestParseProjectDependencies_WalkUpDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	packageJSON := filepath.Join(tmpDir, "package.json")
	subDir := filepath.Join(tmpDir, "src", "components")
	
	// Create a simple package.json in root
	content := `{
		"name": "test",
		"dependencies": {
			"lodash": "^4.17.21"
		}
	}`
	
	if err := os.WriteFile(packageJSON, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}
	
	// Create subdirectory
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	
	// Parse from subdirectory - should walk up and find package.json
	deps, projectType, err := ParseProjectDependencies(subDir)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if projectType != ProjectTypeJavaScript {
		t.Errorf("Expected ProjectTypeJavaScript, got %s", projectType)
	}
	
	if len(deps) != 1 {
		t.Errorf("Expected 1 dependency, got %d", len(deps))
	}
}
