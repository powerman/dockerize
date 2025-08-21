//go:build !windows

package main

// fileURL creates a file:// URL for Unix paths.
func fileURL(path string) string {
	return "file://" + path
}
