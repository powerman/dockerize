//go:build windows

package main

// getTestShellCmd returns appropriate shell command for Windows.
func getTestShellCmd() []string {
	// On Windows, use a simpler approach with powershell that's more reliable
	// Use -NoProfile to avoid loading user profiles that might affect execution
	return []string{"powershell", "-NoProfile", "-Command", "Start-Sleep -Seconds 1; exit 42"}
}

// getTestShellWithSignalCmd returns shell command that handles signals on Windows.
//
//nolint:unused // Used in platform-specific tests.
func getTestShellWithSignalCmd() []string {
	// On Windows, we can't handle Unix signals the same way
	// Just use a simple command that exits with 42
	return []string{"powershell", "-NoProfile", "-Command", "Start-Sleep -Seconds 1; exit 42"}
}
