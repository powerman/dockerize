package main

import (
	"context"
	"strings"
	"testing"

	"github.com/powerman/check"
	"github.com/powerman/gotest/testexec"
)

var ctx = context.Background() // nolint:gochecknoglobals

func TestFlagHelp(tt *testing.T) {
	t := check.T(tt)
	t.Parallel()
	out, err := testexec.Func(ctx, t, main, "-h").CombinedOutput()
	t.Match(err, "exit status 124")
	t.Match(out, "Usage:")
}

func TestFlagVersion(tt *testing.T) {
	t := check.T(tt)
	t.Parallel()
	out, err := testexec.Func(ctx, t, main, "-version").CombinedOutput()
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
		{[]string{"-env", "file:///dev/null"},
			`http/https`},
		{[]string{"-env", "/dev/null"}, ``},
		{[]string{"-env", "http://file.ini"}, ``},
		{[]string{"-env", "https://file.ini"}, ``},
		{[]string{"-env-header", ""},
			`name:value`},
		{[]string{"-env-header", "bad"},
			`name:value`},
		{[]string{"-env-header", " : "},
			`name:value`},
		{[]string{"-env-header", " name : "},
			`name:value`},
		{[]string{"-env-header", " : value "},
			`name:value`},
		{[]string{"-env-header", "n:v", "-env-header", ":", "-env-header", "n:v"},
			`name:value`},
		{[]string{"-env-header", " name : some value "},
			`-env with HTTP`},
		{[]string{"-env-header", "n:v", "-env", "/dev/null"},
			`-env with HTTP`},
		{[]string{"-template", ""},
			`src:dst or src`},
		{[]string{"-template", "a:b:c"},
			`src:dst or src`},
		{[]string{"-template", ":"},
			`src:dst or src`},
		{[]string{"-template", ":b"},
			`src:dst or src`},
		{[]string{"-template", " "}, ``},
		{[]string{"-template", "a", "-template", "a:", "-template", "a:b"}, ``},
		{[]string{"-no-overwrite"},
			`-template`},
		{[]string{"-no-overwrite", "-template", "a"}, ``},
		{[]string{"-delims", ""},
			`left:right`},
		{[]string{"-delims", ":"},
			`left:right`},
		{[]string{"-delims", "a:"},
			`left:right`},
		{[]string{"-delims", ":b"},
			`left:right`},
		{[]string{"-delims", "a a:b"},
			`left:right`},
		{[]string{"-delims", "a:b"},
			`-template`},
		{[]string{"-delims", " a: b ", "-template", "a"}, ``},
		{[]string{"-wait", ""},
			`file/tcp/tcp4/tcp6/unix/http/https`},
		{[]string{"-wait", "/dev/null"},
			`file/tcp/tcp4/tcp6/unix/http/https`},
		{[]string{"-wait", "file:///dev/null", "-wait", "http:", "-wait", "https:"}, ``},
		{[]string{"-wait", "tcp:", "-wait", "tcp4:", "-wait", "tcp6:", "-wait", "unix:"}, ``},
		{[]string{"-wait-http-header", ""},
			`name:value`},
		{[]string{"-wait-http-header", "a:b"},
			`-wait with HTTP`},
		{[]string{"-wait-http-header", "a:b", "-wait", "unix:"},
			`-wait with HTTP`},
		{[]string{"-wait-http-header", "a:b", "-wait", "http:"}, ``},
		{[]string{"-wait-http-header", "a:b", "-wait", "https:"}, ``},
		{[]string{"-skip-tls-verify"},
			`-wait/-env`},
		{[]string{"-skip-tls-verify", "-wait", "unix:"},
			`-wait/-env`},
		{[]string{"-skip-tls-verify", "-env", "http:"}, ``},
		{[]string{"-skip-tls-verify", "-wait", "http:"}, ``},
		{[]string{"-wait-http-skip-redirect"},
			`-wait with HTTP`},
		{[]string{"-wait-http-skip-redirect", "-wait", "unix:"},
			`-wait with HTTP`},
		{[]string{"-wait-http-skip-redirect", "-wait", "http:"}, ``},
		{[]string{"-wait-http-status-code", ""},
			`syntax`},
		{[]string{"-wait-http-status-code", "99"},
			`between 100 and 599`},
		{[]string{"-wait-http-status-code", "600"},
			`between 100 and 599`},
		{[]string{"-wait-http-status-code", "200"},
			`-wait with HTTP`},
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
			out, err := testexec.Func(ctx, t, main, flags...).CombinedOutput()
			if v.want == "" {
				t.Nil(err)
				t.Match(out, ver)
			} else {
				t.Match(err, "exit status 124")
				t.Match(out, `invalid value .* `+v.flags[0]+`:.*`+v.want)
			}
		})
	}
}

func TestFailedINI(tt *testing.T) {
	t := check.T(tt)
	t.Parallel()
	out, err := testexec.Func(ctx, t, main, "-env", "nosuch.ini").CombinedOutput()
	t.Match(err, "exit status 123")
	t.Match(out, `nosuch.ini: no such file`)
}

func TestFailedTemplate(tt *testing.T) {
	t := check.T(tt)
	t.Parallel()
	out, err := testexec.Func(ctx, t, main, "-template", "nosuch.tmpl").CombinedOutput()
	t.Match(err, "exit status 123")
	t.Match(out, `nosuch.tmpl: no such file`)
}

func TestFailedWait(tt *testing.T) {
	t := check.T(tt)
	t.Parallel()
	out, err := testexec.Func(ctx, t, main, "-wait", "file:///nosuch", "-timeout", "0.1s").CombinedOutput()
	t.Match(err, "exit status 123")
	t.Match(out, `/nosuch: no such file`)
}

func TestNothing(tt *testing.T) {
	t := check.T(tt)
	t.Parallel()
	out, err := testexec.Func(ctx, t, main).CombinedOutput()
	t.Nil(err)
	t.Match(out, `^$`)
}
