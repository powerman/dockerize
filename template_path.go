package main

import (
	"net/url"
	"runtime"
	"strings"

	"github.com/powerman/winfilepath"
)

const (
	windowsOS         = "windows"
	colonSplitLimit   = 2
	driveLetterLength = 2
)

// parseTemplatePath parses template path with OS-specific logic.
// Routes to appropriate implementation based on GOOS.
func parseTemplatePath(srcdst string) (src, dst string) {
	if runtime.GOOS == windowsOS {
		return parseTemplatePathWindows(srcdst)
	}
	return parseTemplatePathUnix(srcdst)
}

// parseTemplatePathWindows handles Windows-specific path parsing logic.
// Uses consistent logic with Unix version for multiple colons.
func parseTemplatePathWindows(srcdst string) (src, dst string) {
	colonCount := strings.Count(srcdst, ":")

	// No colons - just source
	if colonCount == 0 {
		return srcdst, ""
	}

	// Check if starts with drive letter first (before counting colons)
	if isDriveLetterPath(srcdst) {
		return parseDriveLetterPath(srcdst, colonCount)
	}

	// Not a drive letter path
	if colonCount == 1 {
		// Single colon - always split
		parts := strings.SplitN(srcdst, ":", colonSplitLimit)
		return parts[0], parts[1]
	}

	// Multiple colons without drive prefix - ERROR (like Unix)
	return "", ""
}

// isDriveLetterPath checks if path starts with a Windows drive letter.
func isDriveLetterPath(srcdst string) bool {
	return len(srcdst) >= driveLetterLength && isLetter(srcdst[0]) && srcdst[1] == ':'
}

// parseDriveLetterPath handles parsing of paths that start with drive letters.
func parseDriveLetterPath(srcdst string, colonCount int) (src, dst string) {
	// Special case: check for patterns like "a:b:c" which should be errors
	if colonCount > 1 && len(srcdst) > driveLetterLength && srcdst[driveLetterLength] != '\\' {
		// Check if there's another colon very close (suggesting "a:b:c" pattern)
		nextColonIdx := strings.Index(srcdst[driveLetterLength:], ":")
		if nextColonIdx != -1 && nextColonIdx <= 3 {
			// Colon too close, likely "a:b:c" pattern - treat as error
			return "", ""
		}
	}

	if colonCount == 1 {
		return parseSingleColonDrivePath(srcdst)
	}

	// Multiple colons with drive letter - find second colon for split
	secondColonIdx := strings.Index(srcdst[driveLetterLength:], ":")
	if secondColonIdx != -1 {
		// Split at the second colon
		splitPos := secondColonIdx + driveLetterLength
		return srcdst[:splitPos], srcdst[splitPos+1:]
	}
	// This shouldn't happen if colonCount > 1, but just in case
	return srcdst, ""
}

// parseSingleColonDrivePath handles single colon in drive letter paths.
func parseSingleColonDrivePath(srcdst string) (src, dst string) {
	// Check if it's just drive letter or drive with path
	if len(srcdst) == driveLetterLength {
		// Just "C:" - no split
		return srcdst, ""
	}
	if len(srcdst) > driveLetterLength && srcdst[driveLetterLength] == '\\' {
		// "C:\\path" - absolute path, no split
		return srcdst, ""
	}
	// "C:temp" - drive + relative path, split after drive letter
	return srcdst[:driveLetterLength-1], srcdst[driveLetterLength:]
}

// parseTemplatePathUnix handles Unix-specific path parsing logic.
func parseTemplatePathUnix(srcdst string) (src, dst string) {
	colonCount := strings.Count(srcdst, ":")

	// No colons - just source
	if colonCount == 0 {
		return srcdst, ""
	}

	// Single colon - always split
	if colonCount == 1 {
		parts := strings.SplitN(srcdst, ":", colonSplitLimit)
		return parts[0], parts[1]
	}

	// Multiple colons on Unix - invalid (like "a:b:c")
	return "", ""
}

// isLetter checks if a character is a letter (A-Z or a-z).
func isLetter(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

// isWindowsDrivePath uses winfilepath to determine if this is a Windows drive path.
func isWindowsDrivePath(path string) bool {
	// For URL paths with forward slashes, convert temporarily for checking
	urlPath := path
	if strings.Contains(path, "/") && !strings.Contains(path, "\\") {
		// This looks like a URL path with forward slashes - convert for checking
		urlPath = strings.ReplaceAll(path, "/", "\\")
	}

	// Use winfilepath.VolumeName to detect Windows drive letters
	vol := winfilepath.VolumeName(urlPath)
	if vol == "" || vol[len(vol)-1] != ':' {
		return false
	}

	// For single letter drives like "C:", check if it's followed by \\ or end of string
	if len(vol) == driveLetterLength && (vol[0] >= 'A' && vol[0] <= 'Z' || vol[0] >= 'a' && vol[0] <= 'z') {
		// Must be followed by \\ or be the whole string
		if len(urlPath) == driveLetterLength || (len(urlPath) > driveLetterLength && urlPath[driveLetterLength] == '\\') {
			return true
		}
	}

	// For UNC paths and other Windows volume formats
	if len(vol) > driveLetterLength {
		return true
	}

	return false
}

// getFilePathFromURL extracts the file path from a file:// URL.
// Uses platform-specific path handling.
func getFilePathFromURL(u *url.URL) string {
	scheme := u.Scheme
	path := u.Path

	if scheme != "file" {
		return path
	}

	// Remove leading slash for Windows paths like /C:/path
	if len(path) >= 3 && path[0] == '/' && isWindowsDrivePath(path[1:]) {
		path = path[1:]
	}

	// Use platform-specific path conversion
	if runtime.GOOS == windowsOS {
		// On Windows, use winfilepath to convert forward slashes to backslashes
		return winfilepath.FromSlash(path)
	}

	// On Unix, return path as-is unless it's a Windows path
	if isWindowsDrivePath(path) {
		// Convert Windows path to use backslashes even on Unix (for testing)
		return winfilepath.FromSlash(path)
	}

	return path
}
