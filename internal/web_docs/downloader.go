package web_docs

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Downloader handles downloading and caching documentation websites using wget
type Downloader struct {
	cacheRoot string
}

// NewDownloader creates a new documentation site downloader
// cacheRoot defaults to ~/.pd/cache if empty
func NewDownloader(cacheRoot string) *Downloader {
	if cacheRoot == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			cacheRoot = ".pd/cache"
		} else {
			cacheRoot = filepath.Join(homeDir, ".pd", "cache")
		}
	}

	return &Downloader{
		cacheRoot: cacheRoot,
	}
}

// DownloadResult contains information about the downloaded documentation
type DownloadResult struct {
	CachePath string // Path to the cached documentation directory
	IndexPath string // Path to the index.html file
}

// Download fetches a documentation website and caches it locally using wget
// Returns the local path to the cached documentation
func (d *Downloader) Download(docURL, langSys, packageName, version string) (*DownloadResult, error) {
	// Create cache directory for this package version
	cacheDir := d.getCachePath(langSys, packageName, version)

	// Check if already cached
	indexPath := filepath.Join(cacheDir, "index.html")
	if _, err := os.Stat(indexPath); err == nil {
		// Already cached
		return &DownloadResult{
			CachePath: cacheDir,
			IndexPath: indexPath,
		}, nil
	}

	// Check if wget is available
	if _, err := exec.LookPath("wget"); err != nil {
		return nil, fmt.Errorf("wget command not found in PATH: %w\nInstall wget (e.g., 'brew install wget' on macOS)", err)
	}

	// Create cache directory
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Use wget to download the site recursively with mirroring
	// --mirror: turn on options suitable for mirroring
	// --convert-links: convert links for offline viewing
	// --page-requisites: download all files necessary to display the page (CSS, images, etc)
	// --no-parent: don't ascend to parent directory
	// --adjust-extension: add .html extension to text/html files
	// --directory-prefix: save to specific directory
	// --no-host-directories: don't create hostname directory
	cmd := exec.Command("wget",
		"--mirror",
		"--convert-links",
		"--page-requisites",
		"--no-parent",
		"--adjust-extension",
		"--directory-prefix="+cacheDir,
		"--no-host-directories",
		"--quiet",
		docURL,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("wget failed: %w\nOutput: %s", err, string(output))
	}

	// Verify index.html exists
	if _, err := os.Stat(indexPath); err != nil {
		return nil, fmt.Errorf("wget completed but index.html not found at %s", indexPath)
	}

	return &DownloadResult{
		CachePath: cacheDir,
		IndexPath: indexPath,
	}, nil
}

// getCachePath returns the cache directory path for a package version
func (d *Downloader) getCachePath(langSys, packageName, version string) string {
	return filepath.Join(d.cacheRoot, langSys, packageName, version, "site")
}

// IsCached checks if documentation for a package version is already cached
func (d *Downloader) IsCached(langSys, packageName, version string) bool {
	indexPath := filepath.Join(d.getCachePath(langSys, packageName, version), "index.html")
	_, err := os.Stat(indexPath)
	return err == nil
}

// GetCachePath returns the cache directory path for a package version
func (d *Downloader) GetCachePath(langSys, packageName, version string) string {
	return d.getCachePath(langSys, packageName, version)
}

// ClearCache removes cached documentation for a specific package version
func (d *Downloader) ClearCache(langSys, packageName, version string) error {
	cachePath := d.getCachePath(langSys, packageName, version)
	return os.RemoveAll(cachePath)
}
