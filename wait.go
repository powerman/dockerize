package main

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

type waitConfig struct {
	headers       httpHeadersFlag
	skipTLSVerify bool
	skipRedirect  bool
	statusCodes   statusCodesFlag
	timeout       time.Duration
	delay         time.Duration
}

func waitForDependencies(cfg waitConfig, urls []*url.URL) { // nolint:gocyclo
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
					ticker := time.NewTicker(cfg.delay)
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
							log.Printf("Problem with check file %s exist: %v. Sleeping %s\n", u.String(), err.Error(), cfg.delay)
						}
					}
				}(u)
			case "tcp", "tcp4", "tcp6":
				wg.Add(1)
				go waitForSocket(cfg, wg, u.Scheme, u.Host)
			case "unix":
				wg.Add(1)
				go waitForSocket(cfg, wg, u.Scheme, u.Path)
			case "http", "https":
				wg.Add(1)
				go func(u *url.URL) {
					transport := &http.Transport{
						TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.skipTLSVerify},
					}
					client := &http.Client{
						Transport: transport,
						Timeout:   cfg.timeout,
					}

					if cfg.skipRedirect {
						client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
							return http.ErrUseLastResponse
						}
					}

					defer wg.Done()
					for {
						req, err := http.NewRequest("GET", u.String(), nil)
						if err != nil {
							log.Printf("Problem with dial: %v. Sleeping %s\n", err.Error(), cfg.delay)
							time.Sleep(cfg.delay)
						}
						if len(cfg.headers) > 0 {
							for _, header := range cfg.headers {
								req.Header.Add(header.name, header.value)
							}
						}

						resp, err := client.Do(req)
						switch {
						case err != nil:
							log.Printf("Problem with request: %s. Sleeping %s\n", err.Error(), cfg.delay)
							time.Sleep(cfg.delay)
						case len(cfg.statusCodes) > 0:
							for _, code := range cfg.statusCodes {
								if code == resp.StatusCode {
									log.Printf("Received %d from %s\n", resp.StatusCode, u.String())
									return
								}
							}
							log.Printf("Received %d from %s. Sleeping %s\n", resp.StatusCode, u.String(), cfg.delay)
							time.Sleep(cfg.delay)
						case err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300:
							log.Printf("Received %d from %s\n", resp.StatusCode, u.String())
							return
						default:
							log.Printf("Received %d from %s. Sleeping %s\n", resp.StatusCode, u.String(), cfg.delay)
							time.Sleep(cfg.delay)
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
	case <-time.After(cfg.timeout):
		// TODO include only timed out dependencies, not all of them
		log.Fatalf("Timeout after %s waiting on dependencies to become available: %v", cfg.timeout, urls)
	}
}

func waitForSocket(cfg waitConfig, wg *sync.WaitGroup, scheme, addr string) {
	defer wg.Done()
	for {
		conn, err := net.DialTimeout(scheme, addr, cfg.timeout)
		if err != nil {
			log.Printf("Problem with dial: %v. Sleeping %s\n", err.Error(), cfg.delay)
			time.Sleep(cfg.delay)
		}
		if conn != nil {
			log.Printf("Connected to %s://%s\n", scheme, addr)
			return
		}
	}
}
