//go:build !windows

package main

import (
	"net/url"
)

// parseFileURL uses standard URL parsing on Unix systems.
func parseFileURL(value string) (*url.URL, error) {
	return url.Parse(value)
}
