//go:build !windows

package main

// getNullDevice returns the null device path for Unix.
//
//nolint:unused // Used in platform-specific tests.
func getNullDevice() string {
	return "/dev/null"
}
