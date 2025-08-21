//go:build windows

package main

import (
	"strings"
)

// fileURL creates a proper file:// URL for Windows paths
func fileURL(path string) string {
	// Convert backslashes to forward slashes for URL
	path = strings.ReplaceAll(path, "\\", "/")
	// For Windows, file URLs should be file:///C:/path format
	if len(path) >= 2 && path[1] == ':' {
		// C:/path -> file:///C:/path
		return "file:///" + path
	}
	return "file://" + path
}
