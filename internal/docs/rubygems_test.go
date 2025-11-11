package docs

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetGemVersion(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rubygems/rails/versions/7.0.0.json" {
			t.Errorf("Expected path /rubygems/rails/versions/7.0.0.json, got %s", r.URL.Path)
		}

		response := GemVersion{
			Number:      "7.0.0",
			PreRelease:  false,
			Platform:    "ruby",
			Summary:     "Full-stack web application framework.",
			Description: "Ruby on Rails is a full-stack web framework...",
			Metadata: GemMetadata{
				Homepage:      "https://rubyonrails.org",
				Documentation: "https://api.rubyonrails.org",
				SourceCode:    "https://github.com/rails/rails",
			},
		}

		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &RubyGemsAPIClient{
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	gemVersion, err := client.GetGemVersion("rails", "7.0.0")
	if err != nil {
		t.Fatalf("GetGemVersion failed: %v", err)
	}

	if gemVersion.Number != "7.0.0" {
		t.Errorf("Expected version 7.0.0, got %s", gemVersion.Number)
	}

	if gemVersion.Metadata.Documentation != "https://api.rubyonrails.org" {
		t.Errorf("Expected documentation URL https://api.rubyonrails.org, got %s", gemVersion.Metadata.Documentation)
	}
}

func TestGetGemVersionNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not found"))
	}))
	defer server.Close()

	client := &RubyGemsAPIClient{
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	_, err := client.GetGemVersion("nonexistent", "1.0.0")
	if err == nil {
		t.Error("Expected error for non-existent gem, got nil")
	}
}

func TestGetGemInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rubygems/rspec.json" {
			t.Errorf("Expected path /rubygems/rspec.json, got %s", r.URL.Path)
		}

		response := GemInfo{
			Name:             "rspec",
			Version:          "3.12.0",
			Authors:          "RSpec Team",
			Info:             "BDD for Ruby",
			DocumentationURI: "https://rspec.info/documentation/",
			HomepageURI:      "https://rspec.info",
			Metadata: GemMetadata{
				Homepage:      "https://rspec.info",
				Documentation: "https://rspec.info/documentation/",
				SourceCode:    "https://github.com/rspec/rspec",
			},
		}

		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &RubyGemsAPIClient{
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	gemInfo, err := client.GetGemInfo("rspec")
	if err != nil {
		t.Fatalf("GetGemInfo failed: %v", err)
	}

	if gemInfo.Name != "rspec" {
		t.Errorf("Expected name rspec, got %s", gemInfo.Name)
	}

	if gemInfo.DocumentationURI != "https://rspec.info/documentation/" {
		t.Errorf("Expected documentation URI https://rspec.info/documentation/, got %s", gemInfo.DocumentationURI)
	}
}

func TestGetDocumentationURL_VersionMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rubygems/rails/versions/7.0.0.json" {
			response := GemVersion{
				Number: "7.0.0",
				Metadata: GemMetadata{
					Documentation: "https://api.rubyonrails.org",
					Homepage:      "https://rubyonrails.org",
				},
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	client := &RubyGemsAPIClient{
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	url, err := client.GetDocumentationURL("rails", "7.0.0")
	if err != nil {
		t.Fatalf("GetDocumentationURL failed: %v", err)
	}

	if url != "https://api.rubyonrails.org" {
		t.Errorf("Expected https://api.rubyonrails.org, got %s", url)
	}
}

func TestGetDocumentationURL_GemInfoFallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rubygems/rails/versions/7.0.0.json" {
			// Version endpoint fails
			w.WriteHeader(http.StatusNotFound)
		} else if r.URL.Path == "/rubygems/rails.json" {
			response := GemInfo{
				Name:             "rails",
				DocumentationURI: "https://guides.rubyonrails.org",
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	client := &RubyGemsAPIClient{
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	url, err := client.GetDocumentationURL("rails", "7.0.0")
	if err != nil {
		t.Fatalf("GetDocumentationURL failed: %v", err)
	}

	if url != "https://guides.rubyonrails.org" {
		t.Errorf("Expected https://guides.rubyonrails.org, got %s", url)
	}
}

func TestGetDocumentationURL_RubydocFallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Both endpoints fail
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := &RubyGemsAPIClient{
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	url, err := client.GetDocumentationURL("mygem", "1.2.3")
	if err != nil {
		t.Fatalf("GetDocumentationURL failed: %v", err)
	}

	expected := "https://www.rubydoc.info/gems/mygem/1.2.3"
	if url != expected {
		t.Errorf("Expected %s, got %s", expected, url)
	}
}

func TestGetDocumentationURL_HomepageFallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rubygems/nokogiri/versions/1.13.0.json" {
			response := GemVersion{
				Number: "1.13.0",
				Metadata: GemMetadata{
					// No documentation URI, but has homepage
					Homepage: "https://nokogiri.org",
				},
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	client := &RubyGemsAPIClient{
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	url, err := client.GetDocumentationURL("nokogiri", "1.13.0")
	if err != nil {
		t.Fatalf("GetDocumentationURL failed: %v", err)
	}

	if url != "https://nokogiri.org" {
		t.Errorf("Expected https://nokogiri.org, got %s", url)
	}
}
