//go:build windows

package main

import (
	"net/url"
	"path/filepath"
)

// parseFileURL handles Windows file paths in file:// URLs correctly
func parseFileURL(value string) (*url.URL, error) {
	u, err := url.Parse(value)
	if err != nil {
		return nil, err
	}

	// Fix Windows paths in file:// URLs
	if u.Scheme == "file" && u.Host != "" && len(u.Host) == 1 {
		// This likely means we have "file://C:\path" where C was interpreted as host
		// Convert it back to proper Windows path
		windowsPath := u.Host + ":" + u.Path
		windowsPath = filepath.FromSlash(windowsPath)
		u.Host = ""
		u.Path = windowsPath
	}

	return u, nil
}
