package main

import (
	"testing"
)

func TestParseTemplatePath(t *testing.T) {
	// Test Unix logic directly
	tests := []struct {
		input   string
		wantSrc string
		wantDst string
	}{
		{"file.tmpl", "file.tmpl", ""},
		{"a:b:c", "", ""}, // Invalid case on Unix
		{"file.tmpl:output.txt", "file.tmpl", "output.txt"},
		{"/unix/path:/unix/output", "/unix/path", "/unix/output"},
	}

	for _, tt := range tests {
		src, dst := parseTemplatePathUnix(tt.input)
		if src != tt.wantSrc || dst != tt.wantDst {
			t.Errorf("parseTemplatePathUnix(%q) = (%q, %q), want (%q, %q)",
				tt.input, src, dst, tt.wantSrc, tt.wantDst)
		}
	}

	// Test Windows logic directly
	windowsTests := []struct {
		input   string
		wantSrc string
		wantDst string
	}{
		{"C:", "C:", ""},
		{"C:\\temp\\file.tmpl", "C:\\temp\\file.tmpl", ""},
		{"C:\\temp\\file.tmpl:D:\\output\\file.txt", "C:\\temp\\file.tmpl", "D:\\output\\file.txt"},
		// New test cases for updated logic
		{"C:temp", "C", "temp"},
		{"file:dest", "file", "dest"},
		{"a:b:c", "", ""}, // Error case like Unix
	}

	for _, tt := range windowsTests {
		src, dst := parseTemplatePathWindows(tt.input)
		if src != tt.wantSrc || dst != tt.wantDst {
			t.Errorf("parseTemplatePathWindows(%q) = (%q, %q), want (%q, %q)",
				tt.input, src, dst, tt.wantSrc, tt.wantDst)
		}
	}
}
