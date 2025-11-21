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
