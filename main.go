package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

var buildVersion = "unknown" // nolint:gochecknoglobals

func main() { // nolint:gocyclo
	var cfg struct {
		version       bool
		ini           iniConfig
		templatePaths stringsFlag // file or file:file or dir or dir:dir
		template      templateConfig
		waitURLs      urlsFlag
		wait          waitConfig
		tailStdout    stringsFlag
		tailStderr    stringsFlag
	}

	flag.BoolVar(&cfg.version, "version", false, "print version and exit")
	flag.StringVar(&cfg.ini.source, "env", "", "path or URL to INI file with default values for unset env vars")
	flag.BoolVar(&cfg.ini.options.AllowPythonMultilineValues, "multiline", false, "allow Python-like multi-line values in INI file")
	flag.StringVar(&cfg.ini.section, "env-section", "", "section name in INI file")
	flag.Var(&cfg.ini.headers, "env-header", "`name:value` or path to file containing name:value for HTTP header to send\n(if -env is an URL)\ncan be passed multiple times")
	flag.Var(&cfg.templatePaths, "template", "template `src:dst` file or dir paths, :dst part is optional\ncan be passed multiple times")
	flag.BoolVar(&cfg.template.noOverwrite, "no-overwrite", false, "do not overwrite existing destination file from template")
	flag.Var(&cfg.template.delims, "delims", "action delimiters in templates")
	flag.Var(&cfg.tailStdout, "stdout", "file `path` to tail to stdout\ncan be passed multiple times")
	flag.Var(&cfg.tailStderr, "stderr", "file `path` to tail to stderr\ncan be passed multiple times")
	flag.Var(&cfg.waitURLs, "wait", "wait for `url` (tcp/tcp4/tcp6/http/https/unix/file)\ncan be passed multiple times")
	flag.Var(&cfg.wait.headers, "wait-http-header", "`name:value` for HTTP header to send\n(if -wait use HTTP)\ncan be passed multiple times")
	flag.BoolVar(&cfg.wait.skipTLSVerify, "skip-tls-verify", false, "skip TLS verification for HTTPS -wait and -env urls")
	flag.BoolVar(&cfg.wait.skipRedirect, "wait-http-skip-redirect", false, "do not follow HTTP redirects\n(if -wait use HTTP)")
	flag.Var(&cfg.wait.statusCodes, "wait-http-status-code", "HTTP status `code` to wait for (2xx by default)\ncan be passed multiple times")
	flag.DurationVar(&cfg.wait.timeout, "timeout", 10*time.Second, "timeout for -wait")
	flag.DurationVar(&cfg.wait.delay, "wait-retry-interval", time.Second, "delay before retrying failed -wait")

	flag.Usage = usage
	flag.Parse()

	cfg.ini.skipTLSVerify = cfg.wait.skipTLSVerify
	if cfg.template.delims[0] == "" {
		cfg.template.delims = [2]string{"{{", "}}"}
	}

	var iniURL, iniHTTP, templatePathBad, waitBadScheme, waitHTTP bool
	if u, err := url.Parse(cfg.ini.source); err == nil && u.IsAbs() {
		iniURL = true
		iniHTTP = u.Scheme == "http" || u.Scheme == "https"
	}
	for _, path := range cfg.templatePaths {
		templatePathBad = templatePathBad || strings.Count(path, ":") > 1
	}
	for _, u := range cfg.waitURLs {
		switch u.Scheme {
		case "http", "https":
			waitHTTP = true
		case "tcp", "tcp4", "tcp6", "unix", "file":
		default:
			waitBadScheme = true
		}
	}
	switch {
	case cfg.version:
		fmt.Println(buildVersion)
		os.Exit(0)
	case flag.NArg() == 0 && flag.NFlag() == 0:
		flag.Usage()
		os.Exit(2)
	case iniURL && !iniHTTP:
		fatalFlagValue("scheme must be http/https", "env", cfg.ini.source)
	case len(cfg.ini.headers) > 0 && !iniHTTP:
		fatalFlagValue("require -env with HTTP url", "env-header", cfg.ini.headers)
	case templatePathBad:
		fatalFlagValue("require src:dst or src", "template", cfg.templatePaths)
	case cfg.template.noOverwrite && len(cfg.templatePaths) == 0:
		fatalFlagValue("require -template", "no-overwrite", cfg.template.noOverwrite)
	case cfg.template.delims[0] != "" && len(cfg.templatePaths) == 0:
		fatalFlagValue("require -template", "delims", cfg.template.delims)
	case waitBadScheme:
		fatalFlagValue("scheme must be http/https/tcp/tcp4/tcp6/unix/file", "wait", cfg.waitURLs)
	case len(cfg.wait.headers) > 0 && !waitHTTP:
		fatalFlagValue("require -wait with HTTP url", "wait-http-header", cfg.wait.headers)
	case len(cfg.wait.statusCodes) > 0 && !waitHTTP:
		fatalFlagValue("require -wait with HTTP url", "wait-http-status-code", cfg.wait.statusCodes)
	case cfg.wait.skipRedirect && !waitHTTP:
		fatalFlagValue("require -wait with HTTP url", "wait-http-skip-redirect", cfg.wait.skipRedirect)
	case cfg.wait.skipTLSVerify && !iniHTTP && !waitHTTP:
		fatalFlagValue("require -wait/-env with HTTP url", "skip-tls-verify", cfg.wait.skipTLSVerify)
	}

	defaultEnv, err := loadINISection(cfg.ini)
	if err != nil {
		log.Fatalf("Failed to load INI: %s.", err)
	}

	setDefaultEnv(defaultEnv)

	cfg.template.data.Env = getEnv()
	err = processTemplatePaths(cfg.template, cfg.templatePaths)
	if err != nil {
		log.Fatalf("Failed to process templates: %s.", err)
	}

	err = waitForURLs(cfg.wait, cfg.waitURLs)
	if err != nil {
		log.Fatalf("Failed to wait: %s.", err)
	}

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // happy linter

	for _, path := range cfg.tailStdout {
		tailFile(ctx, wg, path, os.Stdout)
	}
	for _, path := range cfg.tailStderr {
		tailFile(ctx, wg, path, os.Stderr)
	}

	if flag.NArg() > 0 {
		wg.Add(1)
		go runCmd(ctx, wg, cancel, flag.Arg(0), flag.Args()[1:]...)
	}

	wg.Wait()
}

func usage() {
	fmt.Println(`Usage:
  dockerize options [ command [ arg ... ] ]
  dockerize [ options ] command [ arg ... ]

Utility to simplify running applications in docker containers.

Options:`)
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println(`Example: Generate /etc/nginx/nginx.conf using nginx.tmpl as a template, tail nginx logs, wait for a website to become available on port 8000 and then start nginx.`)
	fmt.Println(`
   dockerize -template nginx.tmpl:/etc/nginx/nginx.conf \
             -stdout /var/log/nginx/access.log \
             -stderr /var/log/nginx/error.log \
             -wait tcp://web:8000 \
             nginx -g 'daemon off;'
	`)
	fmt.Println(`For more information, see https://github.com/powerman/dockerize`)
}
