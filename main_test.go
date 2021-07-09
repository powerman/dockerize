package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
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
	out, err := testexec.Func(testCtx, t, main, "-h").CombinedOutput()
	t.Nil(err)
	t.Match(out, "Usage:")
}

func TestFlagVersion(tt *testing.T) {
	t := check.T(tt)
	t.Parallel()
	out, err := testexec.Func(testCtx, t, main, "-version").CombinedOutput()
	t.Nil(err)
	t.Match(out, ver)
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
		{[]string{"-template", "a:b:c"}, `src:dst or src`},
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
	}
	for _, v := range cases {
		v := v
		t.Run(strings.Join(v.flags, " "), func(tt *testing.T) {
			t := check.T(tt)
			t.Parallel()
			flags := append(v.flags, "-version")
			out, err := testexec.Func(testCtx, t, main, flags...).CombinedOutput()
			if v.want == "" {
				t.Nil(err)
				t.Match(out, ver)
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
	out, err := testexec.Func(testCtx, t, main, "-exit-code", "42", "-env", "nosuch.ini").CombinedOutput()
	t.Match(err, "exit status 42")
	t.Match(out, `nosuch.ini: no such file`)
}

func TestFailedTemplate(tt *testing.T) {
	t := check.T(tt)
	t.Parallel()
	out, err := testexec.Func(testCtx, t, main, "-template", "nosuch.tmpl").CombinedOutput()
	t.Match(err, "exit status 123")
	t.Match(out, `nosuch.tmpl: no such file`)
}

func TestFailedStrictTemplate(tt *testing.T) {
	t := check.T(tt)
	t.Parallel()
	out, err := testexec.Func(testCtx, t, main, "-template", "testdata/src1.tmpl", "-template-strict").CombinedOutput()
	t.Match(err, "exit status 123")
	t.Match(out, `no entry for key "C"`)
}

func TestFailedWait(tt *testing.T) {
	t := check.T(tt)
	t.Parallel()
	out, err := testexec.Func(testCtx, t, main, "-wait", "file:///nosuch", "-timeout", "0.1s").CombinedOutput()
	t.Match(err, "exit status 123")
	t.Match(out, `/nosuch: no such file`)
}

func TestNothing(tt *testing.T) {
	t := check.T(tt)
	t.Parallel()
	out, err := testexec.Func(testCtx, t, main).CombinedOutput()
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
			logf[i] = t.NoErrFile(ioutil.TempFile("", "gotest"))
			logn[i] = logf[i].Name()
			defer os.Remove(logn[i])
			defer logf[i].Close()
		}
	}

	cmd := testexec.Func(testCtx, t, main,
		"-stdout", logn[0], "-stdout", logn[1],
		"-stderr", logn[2], "-stderr", logn[3],
	)
	cmd.Stdout = &bytes.Buffer{}
	cmd.Stderr = &bytes.Buffer{}
	t.Nil(cmd.Start())

	time.Sleep(testSecond)
	for i := range logf {
		t.NoErr(logf[i].Write([]byte(fmt.Sprintf("log%d\n", i))))
	}
	time.Sleep(testSecond)

	t.Nil(cmd.Process.Kill())
	t.Match(cmd.Wait(), `signal: killed`)
	stdout := cmd.Stdout.(*bytes.Buffer).String()
	stderr := cmd.Stderr.(*bytes.Buffer).String()
	t.Match(stdout, `(?m)^log0$`)
	t.Match(stdout, `(?m)^log1$`)
	t.Match(stderr, `(?m)^log2$`)
	t.Match(stderr, `(?m)^log3$`)
}

func TestSmoke1(tt *testing.T) {
	t := checkT(tt)
	t.Parallel()

	var logn, filen, unixn string
	if os.Getenv("GO_WANT_HELPER_PROCESS") == "" { // don't do this again in subprocess
		logf := t.NoErrFile(ioutil.TempFile("", "gotest"))
		logn = logf.Name()
		defer os.Remove(logn)
		defer logf.Close()
		filen = t.TempPath()
		unixn = t.TempPath()
	}

	lnTCP := t.NoErrListen(net.Listen("tcp", "127.0.0.1:0"))
	t.Nil(lnTCP.Close())
	lnTCP4 := t.NoErrListen(net.Listen("tcp4", "127.0.0.1:0"))
	t.Nil(lnTCP4.Close())
	mux := http.NewServeMux()
	ts := httptest.NewUnstartedServer(mux)
	defer ts.Close()

	cmd := testexec.Func(testCtx, t, main,
		"-env", "testdata/env1.ini",
		"-template", "testdata/src1.tmpl",
		"-no-overwrite",
		"-wait", "file://"+filen,
		"-wait", "tcp://"+lnTCP.Addr().String(),
		"-wait", "tcp4://"+lnTCP4.Addr().String(),
		"-wait", "unix://"+unixn,
		"-wait", "http://"+ts.Listener.Addr().String()+"/redirect",
		"-timeout", testSecond.String(),
		"-wait-retry-interval", (testSecond / 10).String(),
		"-stderr", logn,
		"sh", "-c", "sleep 1; exit 42",
	)
	cmd.Stdout = &bytes.Buffer{}
	cmd.Stderr = &bytes.Buffer{}
	t.Nil(cmd.Start())

	time.Sleep(testSecond / 2)
	t.Nil(t.NoErrFile(os.Create(filen)).Close())
	defer os.Remove(filen)
	lnUnix := t.NoErrListen(net.Listen("unix", unixn))
	defer lnUnix.Close()
	lnTCP = t.NoErrListen(net.Listen("tcp", lnTCP.Addr().String()))
	defer lnTCP.Close()
	lnTCP4 = t.NoErrListen(net.Listen("tcp4", lnTCP4.Addr().String()))
	defer lnTCP4.Close()
	var callOK bool
	mux.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ok", http.StatusFound)
	})
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		callOK = true
	})
	ts.Start()

	t.Match(cmd.Wait(), `exit status 42`)
	stdout := cmd.Stdout.(*bytes.Buffer).String()
	stderr := cmd.Stderr.(*bytes.Buffer).String()
	t.Equal(stdout, "A=10 B=20 C=31\n")
	t.Contains(stderr, "Ready:")

	t.True(callOK)
}

func TestSmoke2(tt *testing.T) {
	t := checkT(tt)
	t.Parallel()

	dstDir := t.TempPath()
	if strings.Contains(dstDir, "/gotest") { // protect in case of bug in TempPath
		defer os.RemoveAll(dstDir)
	}
	mux := http.NewServeMux()
	ts := httptest.NewUnstartedServer(mux)
	defer ts.Close()

	cmd := testexec.Func(testCtx, t, main,
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
			trap ''                          HUP INT QUIT ABRT ALRM TERM
			trap 'echo USR; exec >/dev/null' USR1 USR2
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
	t.Nil(cmd.Process.Signal(syscall.SIGUSR1))
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

	t.Equal(parts[1], "USR\n")

	t.True(callINI)
	t.True(callRedirect)

	buf := t.NoErrBuf(ioutil.ReadFile(dstDir + "/abc"))
	t.Equal(string(buf), "A=10 B=20 C=32\n777\n")
	buf = t.NoErrBuf(ioutil.ReadFile(dstDir + "/subdir/func"))
	t.Equal(string(buf), "abc exists\nexample.com\nTrue!False!\nJSON value\n0369\n")
}
