//go:build windows

package main

// getNullDevice returns the null device path for Windows.
//
//nolint:unused // Used in platform-specific tests.
func getNullDevice() string {
	return "NUL"
}
