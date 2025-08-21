//go:build !windows

package main

// getTestShellCmd returns appropriate shell command for Unix.
func getTestShellCmd() []string {
	return []string{"sh", "-c", "sleep 1; exit 42"}
}

// getTestShellWithSignalCmd returns shell command that handles signals on Unix.
//
//nolint:unused // Used in platform-specific tests.
func getTestShellWithSignalCmd() []string {
	return []string{"sh", "-c", `
		exec </dev/null 2>/dev/null
		echo $$
		trap ''                          HUP QUIT ABRT ALRM TERM
		trap 'echo INT; exec >/dev/null' INT
		sleep 10 >/dev/null &
		while ! wait; do :; done
		exit 42
		`}
}
