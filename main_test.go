package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/powerman/check"
	"github.com/powerman/gotest/testexec"
)

func TestFlagHelp(tt *testing.T) {
	t := check.T(tt)
	t.Parallel()
	out, err := testexec.Func(t.Context(), t, main, "-h").CombinedOutput()
	t.Nil(err)
	t.Match(out, "Usage:")
}

func TestGetVersionFromBuildInfo(tt *testing.T) {
	t := check.T(tt)
	t.Parallel()

	// Test with nil buildInfo
	version := getVersionFromBuildInfo(nil)
	t.Equal(version, versionUnknown)

	// Table-driven tests for different version scenarios
	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{
			name:     "empty version",
			version:  "",
			expected: versionUnknown,
		},
		{
			name:     "devel version",
			version:  "(devel)",
			expected: versionUnknown,
		},
		{
			name:     "valid version",
			version:  "v1.2.3",
			expected: "v1.2.3",
		},
		{
			name:     "version without v prefix",
			version:  "1.2.3",
			expected: "1.2.3",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			t := check.T(tt)
			t.Parallel()
			buildInfo := &debug.BuildInfo{
				Main: debug.Module{Version: tc.version},
			}
			result := getVersionFromBuildInfo(buildInfo)
			t.Equal(result, tc.expected)
		})
	}
}

func TestFlagVersion(tt *testing.T) {
	t := check.T(tt)
	t.Parallel()
	out, err := testexec.Func(t.Context(), t, main, "-version").CombinedOutput()
	t.Nil(err)
	t.Match(out, getVersion())
}

func TestFlag(tt *testing.T) {
	t := check.T(tt)
	t.Parallel()
	cases := []struct {
		flags []string
		want  string
	}{
		{[]string{"-env", "file:///dev/null"}, `http/https`},
		{[]string{"-env", "/dev/null"}, ``},
		{[]string{"-env", "http://file.ini"}, ``},
		{[]string{"-env", "https://file.ini"}, ``},
		{[]string{"-env-header", ""}, `name:value`},
		{[]string{"-env-header", "bad"}, `name:value`},
		{[]string{"-env-header", " : "}, `name:value`},
		{[]string{"-env-header", " name : "}, `name:value`},
		{[]string{"-env-header", " : value "}, `name:value`},
		{[]string{"-env-header", "n:v", "-env-header", ":", "-env-header", "n:v"}, `name:value`},
		{[]string{"-env-header", " name : some value "}, `-env with HTTP`},
		{[]string{"-env-header", "n:v", "-env", "/dev/null"}, `-env with HTTP`},
		{[]string{"-template", ""}, `src:dst or src`},
		{[]string{"-template", "aa:bb:cc"}, `src:dst or src`},
		{[]string{"-template", ":"}, `src:dst or src`},
		{[]string{"-template", ":b"}, `src:dst or src`},
		{[]string{"-template", " "}, ``},
		{[]string{"-template", "a", "-template", "a:", "-template", "a:b"}, ``},
		{[]string{"-no-overwrite"}, `-template`},
		{[]string{"-no-overwrite", "-template", "a"}, ``},
		{[]string{"-template-strict"}, `-template`},
		{[]string{"-template-strict", "-template", "a"}, ``},
		{[]string{"-delims", ""}, `left:right`},
		{[]string{"-delims", ":"}, `left:right`},
		{[]string{"-delims", "a:"}, `left:right`},
		{[]string{"-delims", ":b"}, `left:right`},
		{[]string{"-delims", "a a:b"}, `left:right`},
		{[]string{"-delims", "a:b"}, `-template`},
		{[]string{"-delims", " a: b ", "-template", "a"}, ``},
		{[]string{"-wait", ""}, `file/tcp/tcp4/tcp6/unix/http/https/amqp/amqps`},
		{[]string{"-wait", "/dev/null"}, `file/tcp/tcp4/tcp6/unix/http/https/amqp/amqps`},
		{[]string{"-wait", "file:///dev/null", "-wait", "http:", "-wait", "https:"}, ``},
		{[]string{"-wait", "tcp:", "-wait", "tcp4:", "-wait", "tcp6:", "-wait", "unix:"}, ``},
		{[]string{"-wait-list", "tcp: tcp4: http: https: unix: file:"}, ``},
		{[]string{"-wait-http-header", ""}, `name:value`},
		{[]string{"-wait-http-header", "a:b"}, `-wait with HTTP`},
		{[]string{"-wait-http-header", "a:b", "-wait", "unix:"}, `-wait with HTTP`},
		{[]string{"-wait-http-header", "a:b", "-wait", "http:"}, ``},
		{[]string{"-wait-http-header", "a:b", "-wait", "https:"}, ``},
		{[]string{"-skip-tls-verify"}, `-wait/-env`},
		{[]string{"-skip-tls-verify", "-wait", "unix:"}, `-wait/-env`},
		{[]string{"-skip-tls-verify", "-env", "http:"}, ``},
		{[]string{"-skip-tls-verify", "-wait", "http:"}, ``},
		{[]string{"-cacert", "/dev/null"}, `-wait/-env`},
		{[]string{"-cacert", "/dev/null", "-wait", "unix:"}, `-wait/-env`},
		{[]string{"-cacert", "/dev/null", "-env", "http:"}, ``},
		{[]string{"-cacert", "/dev/null", "-wait", "http:"}, ``},
		{[]string{"-wait-http-skip-redirect"}, `-wait with HTTP`},
		{[]string{"-wait-http-skip-redirect", "-wait", "unix:"}, `-wait with HTTP`},
		{[]string{"-wait-http-skip-redirect", "-wait", "http:"}, ``},
		{[]string{"-wait-http-status-code", ""}, `syntax`},
		{[]string{"-wait-http-status-code", "99"}, `between 100 and 599`},
		{[]string{"-wait-http-status-code", "600"}, `between 100 and 599`},
		{[]string{"-wait-http-status-code", "200"}, `-wait with HTTP`},
		{[]string{"-wait-http-status-code", "200", "-wait-http-status-code", "201", "-wait", "http:"}, ``},
		{[]string{"-timeout", "1s"}, ``},
		{[]string{"-wait-retry-interval", "1s"}, ``},
		{[]string{"-stdout", "", "-stdout", " ", "-stderr", "  "}, ``},
		{[]string{"-exec"}, `require command to exec`},
		{[]string{"-exec", "-stdout", "/dev/null"}, `not supported`},
		{[]string{"-exec", "-stderr", "/dev/null"}, `not supported`},
	}
	for _, v := range cases {
		t.Run(strings.Join(v.flags, " "), func(tt *testing.T) {
			t := check.T(tt)
			t.Parallel()
			// Skip tests with invalid Windows filenames because they return
			// different error messages than Unix-like systems not handled
			// by os.IsNotExist() used in flag.go.
			if isWindows {
				for _, flag := range v.flags {
					if flag == ":" || flag == " : " || flag == " name : " || flag == " : value " {
						t.Skip("Skipping test with Windows-incompatible filename")
					}
				}
			}
			flags := append(v.flags, "-version") //nolint:gocritic // By design.
			out, err := testexec.Func(t.Context(), t, main, flags...).CombinedOutput()
			if v.want == "" {
				t.Nil(err)
				t.Match(out, getVersion())
			} else {
				t.Match(err, "exit status 2")
				t.Match(out, `invalid value .* `+v.flags[0]+`:.*`+v.want)
			}
		})
	}
}

func TestFailedINI(tt *testing.T) {
	t := check.T(tt)
	t.Parallel()
	out, err := testexec.Func(t.Context(), t, main, "-exit-code", "42", "-env", "nosuch.ini").CombinedOutput()
	t.Match(err, "exit status 42")
	t.Match(out, `nosuch.ini:.*(no such file|cannot find the file)`)
}

func TestFailedTemplate(tt *testing.T) {
	t := check.T(tt)
	t.Parallel()
	out, err := testexec.Func(t.Context(), t, main, "-template", "nosuch.tmpl").CombinedOutput()
	t.Match(err, "exit status 123")
	t.Match(out, `nosuch.tmpl:.*(no such file|cannot find the file)`)
}

func TestFailedStrictTemplate(tt *testing.T) {
	t := check.T(tt)
	t.Parallel()
	out, err := testexec.Func(t.Context(), t, main, "-template", "testdata/src1.tmpl", "-template-strict").CombinedOutput()
	t.Match(err, "exit status 123")
	t.Match(out, `no entry for key "C"`)
}

func TestFailedWait(tt *testing.T) {
	t := checkT(tt)
	t.Parallel()
	badPath := `/nosuch`
	if isWindows {
		badPath = `C:\nosuch`
	}
	out, err := testexec.Func(t.Context(), t, main, "-wait", t.FileURI(badPath), "-timeout", "0.1s").CombinedOutput()
	t.Match(err, "exit status 123")
	t.Match(out, `/nosuch:.*(no such file|cannot find the file)`)
}

func TestNothing(tt *testing.T) {
	t := check.T(tt)
	t.Parallel()
	out, err := testexec.Func(t.Context(), t, main).CombinedOutput()
	t.Nil(err)
	t.Match(out, `^$`)
}

func TestTail(tt *testing.T) {
	t := checkT(tt)
	t.Parallel()

	var logf [4]*os.File
	var logn [4]string
	if os.Getenv("GO_WANT_HELPER_PROCESS") == "" { // don't do this again in subprocess
		for i := range logf {
			logf[i] = t.NoErrFile(os.CreateTemp(t.TempDir(), "gotest"))
			logn[i] = logf[i].Name()
			defer os.Remove(logn[i]) //nolint:gocritic,revive // By design.
			defer logf[i].Close()    //nolint:gocritic,revive // By design.
		}
	}

	cmd := testexec.Func(t.Context(), t, main,
		"-stdout", logn[0], "-stdout", logn[1],
		"-stderr", logn[2], "-stderr", logn[3],
	)
	cmd.Stdout = &bytes.Buffer{}
	cmd.Stderr = &bytes.Buffer{}
	t.Nil(cmd.Start())

	time.Sleep(testSecond)
	for i := range logf {
		t.NoErr(fmt.Fprintf(logf[i], "log%d\n", i))
	}
	time.Sleep(testSecond)

	t.Nil(cmd.Process.Kill())
	if isWindows {
		cmd.Wait() // No SIGKILL on Windows.
	} else {
		t.Match(cmd.Wait(), `signal: killed`)
	}
	stdout := cmd.Stdout.(*bytes.Buffer).String()
	stderr := cmd.Stderr.(*bytes.Buffer).String()
	t.Match(stdout, `(?m)^log0$`)
	t.Match(stdout, `(?m)^log1$`)
	t.Match(stderr, `(?m)^log2$`)
	t.Match(stderr, `(?m)^log3$`)
}

func TestWaitList(tt *testing.T) {
	t := checkT(tt)
	t.Parallel()

	var logn, filen, fileURI, unixn string
	if os.Getenv("GO_WANT_HELPER_PROCESS") == "" { // don't do this again in subprocess
		logf := t.NoErrFile(os.CreateTemp(t.TempDir(), "gotest"))
		logn = logf.Name()
		defer os.Remove(logn)
		defer logf.Close()
		filen = t.TempPath()
		fileURI = t.FileURI(filen)
		unixn = t.TempPath()
	}

	lnTCP := t.NoErrListen((&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0"))
	t.Nil(lnTCP.Close())
	lnTCP4 := t.NoErrListen((&net.ListenConfig{}).Listen(t.Context(), "tcp4", "127.0.0.1:0"))
	t.Nil(lnTCP4.Close())
	mux := http.NewServeMux()
	ts := httptest.NewUnstartedServer(mux)
	defer ts.Close()

	waitListStr := "tcp://" + lnTCP.Addr().String() + " tcp4://" + lnTCP4.Addr().String()
	if !isWindows {
		waitListStr += " unix://" + unixn
	}
	args := []string{
		"-env", "testdata/env1.ini",
		"-template", "testdata/src1.tmpl",
		"-no-overwrite",
		"-wait", fileURI,
		"-wait-list", waitListStr,
		"-wait", "http://" + ts.Listener.Addr().String() + "/redirect",
		"-timeout", testSecond.String(),
		"-wait-retry-interval", (testSecond / 10).String(),
		"-stderr", logn,
	}
	args = append(args, shellCmd...)
	cmd := testexec.Func(t.Context(), t, main, args...)
	cmd.Stdout = &bytes.Buffer{}
	cmd.Stderr = &bytes.Buffer{}
	t.Nil(cmd.Start())

	time.Sleep(testSecond / 2)
	t.Nil(t.NoErrFile(os.Create(filen)).Close())
	defer os.Remove(filen)
	if !isWindows {
		lnUnix := t.NoErrListen((&net.ListenConfig{}).Listen(t.Context(), "unix", unixn))
		defer lnUnix.Close()
	}
	lnTCP = t.NoErrListen((&net.ListenConfig{}).Listen(t.Context(), "tcp", lnTCP.Addr().String()))
	defer lnTCP.Close()
	lnTCP4 = t.NoErrListen((&net.ListenConfig{}).Listen(t.Context(), "tcp4", lnTCP4.Addr().String()))
	defer lnTCP4.Close()
	var callOK bool
	mux.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ok", http.StatusFound)
	})
	mux.HandleFunc("/ok", func(_ http.ResponseWriter, _ *http.Request) {
		callOK = true
	})
	ts.Start()

	t.Match(cmd.Wait(), `exit status 42`)
	stdout := cmd.Stdout.(*bytes.Buffer).String()
	stderr := cmd.Stderr.(*bytes.Buffer).String()
	t.Equal(strings.TrimSpace(stdout), "A=10 B=20 C=31")
	t.Contains(stderr, "Ready:")

	t.True(callOK)
}

func TestSmoke1(tt *testing.T) {
	t := checkT(tt)
	t.Parallel()

	var logn, filen, fileURI, unixn string
	if os.Getenv("GO_WANT_HELPER_PROCESS") == "" { // don't do this again in subprocess
		logf := t.NoErrFile(os.CreateTemp(t.TempDir(), "gotest"))
		logn = logf.Name()
		defer os.Remove(logn)
		defer logf.Close()
		filen = t.TempPath()
		fileURI = t.FileURI(filen)
		unixn = t.TempPath()
	}

	lnTCP := t.NoErrListen((&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0"))
	t.Nil(lnTCP.Close())
	lnTCP4 := t.NoErrListen((&net.ListenConfig{}).Listen(t.Context(), "tcp4", "127.0.0.1:0"))
	t.Nil(lnTCP4.Close())
	mux := http.NewServeMux()
	ts := httptest.NewUnstartedServer(mux)
	defer ts.Close()

	args := []string{
		"-env", "testdata/env1.ini",
		"-template", "testdata/src1.tmpl",
		"-no-overwrite",
		"-wait", fileURI,
		"-wait", "tcp://" + lnTCP.Addr().String(),
		"-wait", "tcp4://" + lnTCP4.Addr().String(),
		"-wait", "http://" + ts.Listener.Addr().String() + "/redirect",
		"-timeout", testSecond.String(),
		"-wait-retry-interval", (testSecond / 10).String(),
		"-stderr", logn,
	}
	if !isWindows {
		args = append(args, "-wait", "unix://"+unixn)
	}
	args = append(args, shellCmd...)
	cmd := testexec.Func(t.Context(), t, main, args...)
	cmd.Stdout = &bytes.Buffer{}
	cmd.Stderr = &bytes.Buffer{}
	t.Nil(cmd.Start())

	time.Sleep(testSecond / 2)
	t.Nil(t.NoErrFile(os.Create(filen)).Close())
	defer os.Remove(filen)
	if !isWindows {
		lnUnix := t.NoErrListen((&net.ListenConfig{}).Listen(t.Context(), "unix", unixn))
		defer lnUnix.Close()
	}
	lnTCP = t.NoErrListen((&net.ListenConfig{}).Listen(t.Context(), "tcp", lnTCP.Addr().String()))
	defer lnTCP.Close()
	lnTCP4 = t.NoErrListen((&net.ListenConfig{}).Listen(t.Context(), "tcp4", lnTCP4.Addr().String()))
	defer lnTCP4.Close()
	var callOK bool
	mux.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ok", http.StatusFound)
	})
	mux.HandleFunc("/ok", func(_ http.ResponseWriter, _ *http.Request) {
		callOK = true
	})
	ts.Start()

	t.Match(cmd.Wait(), `exit status 42`)
	stdout := cmd.Stdout.(*bytes.Buffer).String()
	stderr := cmd.Stderr.(*bytes.Buffer).String()
	t.Equal(strings.TrimSpace(stdout), "A=10 B=20 C=31")
	t.Contains(stderr, "Ready:")

	t.True(callOK)
}

func TestSmoke2(tt *testing.T) {
	t := checkT(tt)
	t.Parallel()

	if isWindows {
		t.Skip("TestSmoke2 uses Unix-specific features not supported on Windows")
	}

	dstDir := t.TempPath()
	if strings.Contains(dstDir, "/gotest") { // protect in case of bug in TempPath
		defer os.RemoveAll(dstDir)
	}
	mux := http.NewServeMux()
	ts := httptest.NewUnstartedServer(mux)
	defer ts.Close()

	cmd := testexec.Func(t.Context(), t, main,
		"-env", "https://"+ts.Listener.Addr().String()+"/ini",
		"-multiline",
		"-env-section", "Vars",
		"-env-header", "User: env",
		"-env-header", "testdata/secret.hdr",
		"-template", "testdata/src2:"+dstDir,
		"-delims", "<<:>>",
		"-wait", "https://"+ts.Listener.Addr().String()+"/redirect",
		"-wait-http-header", "User: wait",
		"-wait-http-header", "testdata/secret.hdr",
		"-skip-tls-verify",
		"-wait-http-skip-redirect",
		"-wait-http-status-code", "302",
		"-wait-http-status-code", "307",
		"-timeout", testSecond.String(),
		"-wait-retry-interval", (testSecond / 10).String(),
		"sh", "-c", `
			exec </dev/null 2>/dev/null
			echo $$
			trap ''                          HUP QUIT ABRT ALRM TERM
			trap 'echo INT; exec >/dev/null' INT
			sleep 10 >/dev/null &
			while ! wait; do :; done
			exit 42
			`,
	)
	cmd.Stdout = &bytes.Buffer{}
	cmd.Stderr = &bytes.Buffer{}
	t.Nil(cmd.Start())

	time.Sleep(testSecond / 2)
	var callINI, callRedirect bool
	mux.HandleFunc("/ini", func(w http.ResponseWriter, r *http.Request) {
		callINI = true
		t.Equal(r.Header.Get("User"), "env")
		t.Equal(r.Header.Get("Pass"), "Secret")
		f, err := os.Open("testdata/env2.ini")
		t.Nil(err)
		t.NoErr(io.Copy(w, f))
		t.Nil(f.Close())
	})
	mux.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) {
		callRedirect = true
		t.Equal(r.Header.Get("User"), "wait")
		t.Equal(r.Header.Get("Pass"), "Secret")
		http.Redirect(w, r, "/nosuch", http.StatusFound)
	})
	ts.StartTLS()

	time.Sleep(testSecond)
	t.Nil(cmd.Process.Signal(syscall.SIGINT))
	t.Nil(cmd.Process.Signal(syscall.SIGTERM))
	time.Sleep(testSecond)
	t.Nil(cmd.Process.Kill())

	t.Match(cmd.Wait(), `signal: killed`)
	stdout := cmd.Stdout.(*bytes.Buffer).String()
	stderr := cmd.Stderr.(*bytes.Buffer).String()
	t.Log(stderr)
	parts := strings.SplitN(stdout, "\n", 2)
	t.Must(t.Len(parts, 2))

	childPID := t.NoErrInt(strconv.Atoi(parts[0]))
	child, err := os.FindProcess(childPID)
	t.Nil(err)
	time.Sleep(testSecond / 2) // wait for OS cleanup after forwarding SIGKILL to child on Linux
	err = child.Kill()
	if err != nil {
		t.Match(err, `process already finished`)
	}
	t.Nil(child.Release())

	t.Equal(parts[1], "INT\n")

	t.True(callINI)
	t.True(callRedirect)

	buf := t.NoErrBuf(os.ReadFile(dstDir + "/abc"))
	t.Equal(string(buf), "A=10 B=20 C=32\n    777\n")
	buf = t.NoErrBuf(os.ReadFile(dstDir + "/subdir/func"))
	t.Equal(string(buf), "abc exists\nexample.com\nTrue!False!\nJSON value\n0369\n")
}
