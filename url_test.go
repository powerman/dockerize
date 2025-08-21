package main

import (
	"net/url"
	"runtime"
	"testing"
)

type urlTestCase struct {
	urlStr string
	want   string
}

func TestGetFilePathFromURL(t *testing.T) {
	// Platform-specific test cases
	var tests []urlTestCase

	if runtime.GOOS == "windows" {
		// Windows expectations
		tests = []urlTestCase{
			// Unix-style paths should be converted to Windows style
			{"file:///tmp/file.txt", "\\tmp\\file.txt"},
			{"file://host/path", "\\path"},
			{"file:///home/user/file.txt", "\\home\\user\\file.txt"},

			// Windows-style paths
			{"file:///C:/temp/file.txt", "C:\\temp\\file.txt"},
			{"file:///D:/path/to/file", "D:\\path\\to\\file"},

			// Non-file URLs (should return path as-is)
			{"http://example.com/path", "/path"},
			{"https://example.com/path", "/path"},
		}
	} else {
		// Unix expectations (current behavior)
		tests = []urlTestCase{
			// Basic file URLs
			{"file:///tmp/file.txt", "/tmp/file.txt"},
			{"file://host/path", "/path"},

			// Windows-style paths (should work on Linux now)
			{"file:///C:/temp/file.txt", "C:\\temp\\file.txt"},
			{"file:///D:/path/to/file", "D:\\path\\to\\file"},

			// Unix-style paths
			{"file:///home/user/file.txt", "/home/user/file.txt"},

			// Non-file URLs (should return path as-is)
			{"http://example.com/path", "/path"},
			{"https://example.com/path", "/path"},
		}
	}

	for _, tt := range tests {
		u, err := url.Parse(tt.urlStr)
		if err != nil {
			t.Fatalf("Failed to parse URL %q: %v", tt.urlStr, err)
		}
		got := getFilePathFromURL(u)
		if got != tt.want {
			t.Errorf("getFilePathFromURL(%q) = %q, want %q", tt.urlStr, got, tt.want)
		}
	}
}
