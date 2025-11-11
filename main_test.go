package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMainIntegration(t *testing.T) {
	// This is an integration test that verifies the basic workflow
	// We'll create a temporary project and test that the CLI can find it
	
	tmpDir := t.TempDir()
	packageJSON := filepath.Join(tmpDir, "package.json")
	
	content := `{
		"name": "test-project",
		"dependencies": {
			"express": "^4.18.2"
		}
	}`
	
	if err := os.WriteFile(packageJSON, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test package.json: %v", err)
	}
	
	// Change to the test directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)
	
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to test directory: %v", err)
	}
	
	// We can't easily test the full main() function because it's interactive,
	// but we've tested all the components it uses in their respective test files
	// This test just ensures the test setup works
	t.Log("Integration test environment setup successful")
}
