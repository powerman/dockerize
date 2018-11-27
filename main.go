package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/ini.v1"
)

const defaultWaitRetryInterval = time.Second

var buildVersion = "unknown" // nolint:gochecknoglobals

type sliceVar []string

// Context is the type passed into the template renderer
type Context struct{}

// HTTPHeader this is an optional header passed on http checks
type HTTPHeader struct {
	name  string
	value string
}

// Env is bound to the template rendering Context and returns the
// environment variables passed to the program
func (c *Context) Env() map[string]string {
	env := make(map[string]string)
	for _, i := range os.Environ() {
		sep := strings.Index(i, "=")
		env[i[0:sep]] = i[sep+1:]
	}
	return env
}

func (s *sliceVar) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func (s *sliceVar) String() string {
	return strings.Join(*s, ",")
}

func waitForDependencies( // nolint:gocyclo
	urls []*url.URL,
	headers []HTTPHeader,
	skipTLSVerify bool,
	skipRedirect bool,
	statusCodes []string,
	timeout time.Duration,
	delay time.Duration,
) {
	wg := &sync.WaitGroup{}
	dependencyChan := make(chan struct{})

	go func() {
		for _, u := range urls {
			log.Println("Waiting for:", u.String())

			switch u.Scheme {
			case "file":
				wg.Add(1)
				go func(u *url.URL) {
					defer wg.Done()
					ticker := time.NewTicker(delay)
					defer ticker.Stop()
					var err error
					for range ticker.C {
						_, err = os.Stat(u.Path)
						switch {
						case err == nil:
							log.Printf("File %s had been generated\n", u.String())
							return
						case os.IsNotExist(err):
							continue
						default:
							log.Printf("Problem with check file %s exist: %v. Sleeping %s\n", u.String(), err.Error(), delay)
						}
					}
				}(u)
			case "tcp", "tcp4", "tcp6":
				wg.Add(1)
				go waitForSocket(wg, u.Scheme, u.Host, timeout, delay)
			case "unix":
				wg.Add(1)
				go waitForSocket(wg, u.Scheme, u.Path, timeout, delay)
			case "http", "https":
				wg.Add(1)
				go func(u *url.URL) {
					transport := &http.Transport{
						TLSClientConfig: &tls.Config{InsecureSkipVerify: skipTLSVerify},
					}
					client := &http.Client{
						Transport: transport,
						Timeout:   timeout,
					}

					if skipRedirect {
						client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
							return http.ErrUseLastResponse
						}
					}

					defer wg.Done()
					for {
						req, err := http.NewRequest("GET", u.String(), nil)
						if err != nil {
							log.Printf("Problem with dial: %v. Sleeping %s\n", err.Error(), delay)
							time.Sleep(delay)
						}
						if len(headers) > 0 {
							for _, header := range headers {
								req.Header.Add(header.name, header.value)
							}
						}

						resp, err := client.Do(req)
						switch {
						case err != nil:
							log.Printf("Problem with request: %s. Sleeping %s\n", err.Error(), delay)
							time.Sleep(delay)
						case len(statusCodes) > 0:
							for _, code := range statusCodes {
								if code == strconv.Itoa(resp.StatusCode) {
									log.Printf("Received %d from %s\n", resp.StatusCode, u.String())
									return
								}
							}
							log.Printf("Received %d from %s. Sleeping %s\n", resp.StatusCode, u.String(), delay)
							time.Sleep(delay)
						case err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300:
							log.Printf("Received %d from %s\n", resp.StatusCode, u.String())
							return
						default:
							log.Printf("Received %d from %s. Sleeping %s\n", resp.StatusCode, u.String(), delay)
							time.Sleep(delay)
						}
					}
				}(u)
			default:
				log.Fatalf("invalid host protocol provided: %s. supported protocols are: tcp, tcp4, tcp6 and http", u.Scheme)
			}
		}
		wg.Wait()
		close(dependencyChan)
	}()

	select {
	case <-dependencyChan:
		break
	case <-time.After(timeout):
		// TODO include only timed out dependencies, not all of them
		log.Fatalf("Timeout after %s waiting on dependencies to become available: %v", timeout, urls)
	}
}

func waitForSocket(wg *sync.WaitGroup, scheme, addr string, timeout, delay time.Duration) {
	defer wg.Done()
	for {
		conn, err := net.DialTimeout(scheme, addr, timeout)
		if err != nil {
			log.Printf("Problem with dial: %v. Sleeping %s\n", err.Error(), delay)
			time.Sleep(delay)
		}
		if conn != nil {
			log.Printf("Connected to %s://%s\n", scheme, addr)
			return
		}
	}
}

func usage() {
	println(`Usage: dockerize [options] [command]

Utility to simplify running applications in docker containers

Options:`)
	flag.PrintDefaults()

	println(`
Arguments:
  command - command to be executed
  `)

	println(`Examples:
`)
	println(`   Generate /etc/nginx/nginx.conf using nginx.tmpl as a template, tail /var/log/nginx/access.log
   and /var/log/nginx/error.log, waiting for a website to become available on port 8000 and start nginx.`)
	println(`
   dockerize -template nginx.tmpl:/etc/nginx/nginx.conf \
             -stdout /var/log/nginx/access.log \
             -stderr /var/log/nginx/error.log \
             -wait tcp://web:8000 nginx
	`)

	println(`For more information, see https://github.com/powerman/dockerize`)
}

func getINI(env string, envHdr []string, skipTLSVerify bool) (iniFile []byte, err error) {
	// See if envFlag parses like an absolute URL, if so use http, otherwise treat as filename
	url, urlERR := url.ParseRequestURI(env)
	if urlERR == nil && url.IsAbs() {
		var resp *http.Response
		var req *http.Request
		var hdr string
		var client *http.Client
		// Define redirect handler to disallow redirects
		var redir = func(req *http.Request, via []*http.Request) error {
			return errors.New("redirects disallowed")
		}

		transport := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: skipTLSVerify},
		}
		client = &http.Client{Transport: transport, CheckRedirect: redir}
		req, err = http.NewRequest("GET", env, nil)
		if err != nil {
			// Weird problem with declaring client, bail
			return iniFile, err
		}
		// Handle headers for request - are they headers or filepaths?
		for _, h := range envHdr {
			if strings.Contains(h, ":") {
				// This will break if path includes colon - don't use colons in path!
				hdr = h
			} else { // Treat this is a path to a secrets file containing header
				var hdrFile []byte
				hdrFile, err = ioutil.ReadFile(h)
				if err != nil { // Could not read file, error out
					return iniFile, err
				}
				hdr = string(hdrFile)
			}
			parts := strings.Split(hdr, ":")
			if len(parts) != 2 {
				log.Fatalf("Bad env-headers argument: %s. expected \"headerName: headerValue\"", hdr)
			}
			req.Header.Add(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
		resp, err = client.Do(req)
		if err == nil && resp.StatusCode == 200 {
			defer resp.Body.Close()
			iniFile, err = ioutil.ReadAll(resp.Body)
		} else if err == nil { // Request completed with unexpected HTTP status code, bail
			err = errors.New(resp.Status)
			return iniFile, err
		}
	} else {
		iniFile, err = ioutil.ReadFile(env)
	}
	return iniFile, err
}

func main() { // nolint:gocyclo
	var cfg struct { // nolint:maligned
		version           bool
		env               string
		multiline         bool
		envSection        string
		envHdr            sliceVar
		templates         sliceVar
		noOverwrite       bool
		stdoutTail        sliceVar
		stderrTail        sliceVar
		delims            string
		headers           sliceVar
		statusCodes       sliceVar
		skipRedirect      bool
		wait              sliceVar
		skipTLSVerify     bool
		waitTimeout       time.Duration
		waitRetryInterval time.Duration
	}

	flag.BoolVar(&cfg.version, "version", false, "show version")
	flag.StringVar(&cfg.env, "env", "", "Optional path to INI file for injecting env vars. Does not overwrite existing env vars")
	flag.BoolVar(&cfg.multiline, "multiline", false, "enable parsing multiline INI entries in INI environment file")
	flag.StringVar(&cfg.envSection, "env-section", "", "Optional section of INI file to use for loading env vars. Defaults to \"\"")
	flag.Var(&cfg.envHdr, "env-header", "Optional string or path to secrets file for http headers passed if -env is a URL")
	flag.Var(&cfg.templates, "template", "Template (/template:/dest). Can be passed multiple times. Does also support directories")
	flag.BoolVar(&cfg.noOverwrite, "no-overwrite", false, "Do not overwrite destination file if it already exists.")
	flag.Var(&cfg.stdoutTail, "stdout", "Tails a file to stdout. Can be passed multiple times")
	flag.Var(&cfg.stderrTail, "stderr", "Tails a file to stderr. Can be passed multiple times")
	flag.StringVar(&cfg.delims, "delims", "", `template tag delimiters. default "{{":"}}" `)
	flag.Var(&cfg.headers, "wait-http-header", "HTTP headers, colon separated. e.g \"Accept-Encoding: gzip\". Can be passed multiple times")
	flag.Var(&cfg.statusCodes, "wait-http-status-code", "HTTP code to wait for e.g. \"-wait-http-status-code 302  -wait-http-status-code 200\". Can be passed multiple times. (If not specified -wait returns on 200 >= x < 300) ")
	flag.BoolVar(&cfg.skipRedirect, "wait-http-skip-redirect", false, "Skip HTTP redirects")
	flag.Var(&cfg.wait, "wait", "Host (tcp/tcp4/tcp6/http/https/unix/file) to wait for before this container starts. Can be passed multiple times. e.g. tcp://db:5432")
	flag.BoolVar(&cfg.skipTLSVerify, "skip-tls-verify", false, "Skip tls verification for https wait requests")
	flag.DurationVar(&cfg.waitTimeout, "timeout", 10*time.Second, "Host wait timeout")
	flag.DurationVar(&cfg.waitRetryInterval, "wait-retry-interval", defaultWaitRetryInterval, "Duration to wait before retrying")

	flag.Usage = usage
	flag.Parse()

	if cfg.version {
		fmt.Println(buildVersion)
		return
	}

	if flag.NArg() == 0 && flag.NFlag() == 0 {
		usage()
		os.Exit(1)
	}

	if cfg.env != "" {
		iniFile, err := getINI(cfg.env, cfg.envHdr, cfg.skipTLSVerify)
		if err != nil {
			log.Fatalf("unreadable INI file %s: %s", cfg.env, err)
		}
		config, err := ini.LoadSources(ini.LoadOptions{AllowPythonMultilineValues: cfg.multiline}, iniFile)
		if err != nil {
			log.Fatalf("error parsing contents of %s as INI format: %s", cfg.env, err)
		}
		envHash := config.Section(cfg.envSection).KeysHash()

		for k, v := range envHash {
			if _, ok := os.LookupEnv(k); !ok {
				// log.Printf("Setting %s to %s", k, v)
				os.Setenv(k, v)
			}
		}
	}

	var delims []string
	if cfg.delims != "" {
		delims = strings.Split(cfg.delims, ":")
		if len(delims) != 2 {
			log.Fatalf("bad delimiters argument: %s. expected \"left:right\"", cfg.delims)
		}
	}

	urls := make([]*url.URL, len(cfg.wait))
	for i, host := range cfg.wait {
		u, err := url.Parse(host)
		if err != nil {
			log.Fatalf("bad hostname provided: %s. %s", host, err.Error())
		}
		urls[i] = u
	}

	var headers []HTTPHeader
	for _, h := range cfg.headers {
		//validate headers need -wait options
		if len(cfg.wait) == 0 {
			log.Fatalf("-wait-http-header \"%s\" provided with no -wait option", h)
		}

		const errMsg = "bad HTTP Headers argument: %s. expected \"headerName: headerValue\""
		if strings.Contains(h, ":") {
			parts := strings.Split(h, ":")
			if len(parts) != 2 {
				log.Fatalf(errMsg, cfg.headers)
			}
			headers = append(headers, HTTPHeader{name: strings.TrimSpace(parts[0]), value: strings.TrimSpace(parts[1])})
		} else {
			log.Fatalf(errMsg, cfg.headers)
		}
	}

	for _, t := range cfg.templates {
		template, dest := t, ""
		if strings.Contains(t, ":") {
			parts := strings.Split(t, ":")
			if len(parts) != 2 {
				log.Fatalf("bad template argument: %s. expected \"/template:/dest\"", t)
			}
			template, dest = parts[0], parts[1]
		}

		fi, err := os.Stat(template)
		if err != nil {
			log.Fatalf("unable to stat %s, error: %s", template, err)
		}
		if fi.IsDir() {
			generateDir(delims, cfg.noOverwrite, template, dest)
		} else {
			generateFile(delims, cfg.noOverwrite, template, dest)
		}
	}

	waitForDependencies(urls, headers, cfg.skipTLSVerify, cfg.skipRedirect,
		cfg.statusCodes, cfg.waitTimeout, cfg.waitRetryInterval)

	ctx := context.Background()
	wg := &sync.WaitGroup{}

	if flag.NArg() > 0 {
		wg.Add(1)
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)
		go runCmd(ctx, wg, cancel, flag.Arg(0), flag.Args()[1:]...)
	}

	for _, path := range cfg.stdoutTail {
		wg.Add(1)
		go tailFile(ctx, wg, path, os.Stdout)
	}

	for _, path := range cfg.stderrTail {
		wg.Add(1)
		go tailFile(ctx, wg, path, os.Stderr)
	}

	wg.Wait()
}
