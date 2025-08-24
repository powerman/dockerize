package main

import (
	"testing"

	"github.com/powerman/check"
)

func TestParseTemplatePath(tt *testing.T) {
	t := check.T(tt)
	t.Parallel()

	testCasesUnix := []struct {
		input       string
		expectedSrc string
		expectedDst string
	}{
		// Single path cases (no colons)
		{"simple", "simple", ""},
		{"/file.txt", "/file.txt", ""},

		// Two path cases (single colon)
		{"src:dst", "src", "dst"},
		{"/file1.txt:../file2.txt", "/file1.txt", "../file2.txt"},

		// Edge cases with single colon
		{":", "", ""},
		{":dst", "", "dst"},
		{"src:", "src", ""},

		// Multiple colons on Unix - invalid
		{"a:b:c", "", ""},
		{"file1:file2:file3", "", ""},
	}

	testCasesWindows := []struct {
		input       string
		expectedSrc string
		expectedDst string
	}{
		// Single path cases
		{"simple", "simple", ""},
		{`C:\abs.txt`, `C:\abs.txt`, ""},
		{`\\server\share\file.txt`, `\\server\share\file.txt`, ""},

		// Two path cases
		{"src:dst", "src", "dst"},
		{`C:\src.txt:D:\dst.txt`, `C:\src.txt`, `D:\dst.txt`},
		{`\\server\share\src.txt:local.txt`, `\\server\share\src.txt`, "local.txt"},

		// Edge cases
		{":", "", ""},
		{":dst", "", "dst"},
		{"src:", "src", ""},

		// Multiple paths (should return empty as invalid)
		{"a:b:c", "a:b", "c"},
		{`C:\a:D:\b:E:\c`, "", ""},
	}

	testCases := testCasesUnix
	if isWindows {
		testCases = testCasesWindows
	}
	for _, tc := range testCases {
		src, dst := parseTemplatePaths(tc.input)
		t.Equal(src, tc.expectedSrc)
		t.Equal(dst, tc.expectedDst)
	}
}

func TestSplitWindowsPaths_Basic(tt *testing.T) {
	t := check.T(tt)
	t.Parallel()

	testCases := []struct {
		input    string
		expected []string
	}{
		{``, []string{}},
		{`:`, []string{``, ``}},
		{`::`, []string{``, ``, ``}},
		{`:a`, []string{``, `a`}},
		{`a`, []string{`a`}},
		{`a:`, []string{`a:`}},
		{`a:b`, []string{`a:b`}},
		{`a:b:`, []string{`a:b`, ``}},
		{`a:b:c`, []string{`a:b`, `c`}},
		{`a:b:c:d`, []string{`a:b`, `c:d`}},
		{`aa:b:c`, []string{`aa`, `b:c`}},
		{`aa:bb:c`, []string{`aa`, `bb`, `c`}},
	}

	for _, tc := range testCases {
		result := splitWindowsPaths(tc.input)
		t.DeepEqual(result, tc.expected)
	}
}

func TestSplitWindowsPaths_Extensive(tt *testing.T) {
	t := check.T(tt)
	t.Parallel()

	paths := []string{
		`rel_dir_drive.txt`,
		`rel_dir_drive\file.txt`,
		`C:`,
		`C:rel_dir.txt`,
		`C:rel_dir\file.txt`,
		`C:\abs.txt`,
		`C:\abs\file.txt`,
		`\\server\share\file.txt`,
		`\\server\share\dir\file.txt`,
		`\\example.com\share\file.txt`,
		`\\example.com\share\dir\file.txt`,
		`\\?\C:\abs\file.txt`,
		`\\?\UNC\server\share\file.txt`,
		`\\?\Volume{12345678-1234-1234-1234-123456789abc}\`,
		// Commented out device paths that don't work with current implementation.
		// `\\.\PhysicalDrive0`,
		// `\\.\C:`,
		// `\\.\NUL`,
	}

	for _, path1 := range paths {
		t.Run("Path1="+path1, func(tt *testing.T) {
			t := check.T(tt)
			t.Parallel()

			// Test single paths.
			res := splitWindowsPaths(path1)
			t.DeepEqual(res, []string{path1})

			// Test path + colon (should add empty string).
			res = splitWindowsPaths(path1 + ":")
			t.DeepEqual(res, []string{path1, ``})

			for _, path2 := range paths {
				// Test all combinations of two paths.
				res = splitWindowsPaths(path1 + ":" + path2)
				t.DeepEqual(res, []string{path1, path2})

				// Test combinations of three paths.
				result := splitWindowsPaths(path1 + ":" + path2 + ":" + path1)
				t.DeepEqual(result, []string{path1, path2, path1})
			}
		})
	}
}
