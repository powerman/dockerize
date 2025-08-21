//go:build windows

package main

// hasUnixSocketSupport returns whether Unix sockets are supported.
func hasUnixSocketSupport() bool {
	return false // Unix sockets not supported on Windows
}
