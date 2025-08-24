package main

import (
	"runtime"
	"strings"

	"github.com/powerman/winfilepath"
)

// parseTemplatePaths parses colon-separated template paths with OS-specific logic.
// On POSIX colons are always separators (so files with colon in name are not supported).
// On Windows colons in drive letters are not separators.
//
// If only one path is given, the dst is empty.
// If two paths are given, the first is src and the second is dst.
// If none or more than two paths are given, both src and dst are empty (this means error).
func parseTemplatePaths(srcdst string) (src, dst string) {
	var paths []string
	if runtime.GOOS == "windows" {
		paths = splitWindowsPaths(srcdst)
	} else {
		paths = strings.Split(srcdst, ":")
	}

	switch len(paths) {
	case 1:
		return paths[0], ""
	case 2: //nolint:mnd // Clearer than a named constant.
		return paths[0], paths[1]
	}
	return "", ""
}

// splitWindowsPaths splits a PATH-like colon-separated string into its components,
// taking into account that Windows drive letters contain colons.
//
// Device names (like `\\.\NUL`) are not supported yet.
func splitWindowsPaths(s string) []string {
	if s == "" {
		return []string{}
	}

	var paths []string //nolint:prealloc // Premature optimization.
	start := 0

	for i := range len(s) {
		if s[i] != ':' {
			continue
		}

		// Check if this colon is part of a Windows volume name.
		currentSegment := s[start : i+1]
		volumeName := winfilepath.VolumeName(currentSegment)
		if volumeName != "" && volumeName == currentSegment {
			continue
		}

		// Otherwise this is a path separator.
		paths = append(paths, s[start:i])
		start = i + 1
	}

	if start <= len(s) {
		paths = append(paths, s[start:])
	}

	return paths
}
