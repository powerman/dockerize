//go:build windows

package main

import (
	"os"
)

// applyOwnership is a no-op on Windows.
func applyOwnership(*os.File, os.FileInfo) error {
	return nil
}

// applyDirOwnership is a no-op on Windows.
func applyDirOwnership(string, os.FileInfo) error {
	return nil
}
