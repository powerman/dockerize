//go:build windows

package main

import (
	"os"
	"strings"
	"testing"

	"github.com/powerman/check"
	"github.com/powerman/gotest/testexec"
)

func TestWindowsBasicExitCode(tt *testing.T) {
	t := check.T(tt)
	t.Parallel()

	// Test basic functionality: run a simple command that should exit with code 42
	out, err := testexec.Func(t.Context(), t, main, "cmd", "/c", "exit 42").CombinedOutput()
	t.Match(err, "exit status 42")
	t.Equal(string(out), "")
}

func TestWindowsShellCommand(tt *testing.T) {
	t := check.T(tt)
	t.Parallel()

	// Test the shell command used in our tests
	shellCmd := getTestShellCmd()
	args := append([]string{"-timeout", "5s"}, shellCmd...)
	out, err := testexec.Func(t.Context(), t, main, args...).CombinedOutput()
	t.Match(err, "exit status 42")
	t.Log("Shell command:", shellCmd)
	t.Log("Shell command output:", string(out))
}

func TestWindowsTemplateProcessing(tt *testing.T) {
	t := checkT(tt)
	// Cannot use t.Parallel() with t.Setenv()

	// Test template processing with Windows paths
	tmpDir := tt.TempDir()
	tmplFile := tmpDir + "\\test.tmpl"
	outFile := tmpDir + "\\test.out"

	// Create a simple template
	t.Helper()
	err := os.WriteFile(tmplFile, []byte("Value: {{.Env.TEMP_VAR}}\n"), 0o644)
	t.Nil(err)

	// Set environment variable
	tt.Setenv("TEMP_VAR", "test_value")

	// Process template - use the parseTemplatePath format
	templateArg := tmplFile + ":" + outFile
	out, err := testexec.Func(t.Context(), t, main, "-template", templateArg).CombinedOutput()
	t.Nil(err)
	t.Equal(string(out), "")

	// Check output
	content := t.NoErrBuf(os.ReadFile(outFile))
	t.Equal(strings.TrimSpace(string(content)), "Value: test_value")
}

func TestWindowsFileURLHandling(tt *testing.T) {
	t := checkT(tt)
	t.Parallel()

	// Test file URL handling on Windows
	tmpDir := tt.TempDir()
	testFile := tmpDir + "\\test.txt"

	// Create a test file
	err := os.WriteFile(testFile, []byte("test content"), 0o644)
	t.Nil(err)

	// Test file:// URL
	fileURL := fileURL(testFile)
	t.Log("File URL:", fileURL)
	t.Match(fileURL, "^file://")

	// Test waiting for file
	out, err := testexec.Func(t.Context(), t, main, "-wait", fileURL, "-timeout", "1s", "cmd", "/c", "echo done").CombinedOutput()
	t.Nil(err)
	t.Match(string(out), "done")
}
