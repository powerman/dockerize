package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	errUnexpectedStatusCode = errors.New("unexpected HTTP status code")
	errSchemeNotSupported   = errors.New("wait scheme not supported")
	errTimedOut             = errors.New("timed out")
)

type waitConfig struct {
	headers       httpHeadersFlag
	skipTLSVerify bool
	ca            *x509.CertPool
	skipRedirect  bool
	statusCodes   statusCodesFlag
	timeout       time.Duration
	delay         time.Duration
}

func waitForURLs(cfg waitConfig, urls []*url.URL) error {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.timeout)
	defer cancel()

	waiting := make(map[*url.URL]bool, len(urls))
	readyc := make(chan *url.URL, len(urls))
	for _, u := range urls {
		if !waiting[u] { // skip possible duplicates
			waiting[u] = true
			switch u.Scheme {
			case schemeFile:
				go waitForPath(ctx, cfg, u, readyc)
			case schemeTCP, schemeTCP4, schemeTCP6, schemeUnix:
				go waitForSocket(ctx, cfg, u, readyc)
			case schemeHTTP, schemeHTTPS:
				go waitForHTTP(ctx, cfg, u, readyc)
			case schemeAMQP, schemeAMQPS:
				go waitForAMQP(ctx, cfg, u, readyc)
			default:
				return fmt.Errorf("%w: %s", errSchemeNotSupported, u)
			}
		}
	}

	for len(waiting) > 0 {
		select {
		case u := <-readyc:
			log.Printf("Ready: %s.", u)
			delete(waiting, u)
		case <-ctx.Done():
			for s := range waiting {
				return fmt.Errorf("%w: %s", errTimedOut, s)
			}
			panic("internal error")
		}
	}
	return nil
}

func waitForPath(ctx context.Context, cfg waitConfig, u *url.URL, readyc chan<- *url.URL) {
	for {
		_, err := os.Stat(u.Path)
		if err == nil {
			break
		}
		log.Printf("Waiting for %s: %s.", u, err)
		select {
		case <-time.After(cfg.delay):
		case <-ctx.Done():
			return
		}
	}

	readyc <- u
}

func waitForSocket(ctx context.Context, cfg waitConfig, u *url.URL, readyc chan<- *url.URL) {
	addr := u.Host
	if u.Scheme == schemeUnix {
		addr = u.Path
	}
	dialer := &net.Dialer{}

	for {
		conn, err := dialer.DialContext(ctx, u.Scheme, addr)
		if err == nil {
			warnIfFail(conn.Close)
			break
		}
		log.Printf("Waiting for %s: %s.", u, err)
		select {
		case <-time.After(cfg.delay):
		case <-ctx.Done():
			return
		}
	}

	readyc <- u
}

func waitForHTTP(ctx context.Context, cfg waitConfig, u *url.URL, readyc chan<- *url.URL) {
	waitStatusCode := make(map[int]bool)
	if len(cfg.statusCodes) == 0 {
		for statusCode := http.StatusOK; statusCode < http.StatusMultipleChoices; statusCode++ {
			waitStatusCode[statusCode] = true
		}
	} else {
		for _, statusCode := range cfg.statusCodes {
			waitStatusCode[statusCode] = true
		}
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: cfg.skipTLSVerify, //nolint:gosec // TLS InsecureSkipVerify may be true.
				RootCAs:            cfg.ca,
			},
		},
	}
	if cfg.skipRedirect {
		client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	var resp *http.Response

	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), http.NoBody)
		if err == nil {
			for _, h := range cfg.headers {
				req.Header.Add(h.name, h.value)
			}
			resp, err = client.Do(req.WithContext(ctx))
		}
		if err == nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			if waitStatusCode[resp.StatusCode] {
				break
			}
			err = fmt.Errorf("%w: %d", errUnexpectedStatusCode, resp.StatusCode)
		}
		log.Printf("Waiting for %s: %s.", u, err)
		select {
		case <-time.After(cfg.delay):
		case <-ctx.Done():
			return
		}
	}

	readyc <- u
}

func waitForAMQP(ctx context.Context, cfg waitConfig, u *url.URL, readyc chan<- *url.URL) {
	amqpCfg := amqp.Config{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.skipTLSVerify, //nolint:gosec // TLS InsecureSkipVerify may be true.
			RootCAs:            cfg.ca,
		},
	}

	for {
		if deadline, ok := ctx.Deadline(); ok {
			amqpCfg.Dial = amqp.DefaultDial(time.Until(deadline))
		}
		conn, err := amqp.DialConfig(u.String(), amqpCfg)
		if err == nil {
			_, err = conn.Channel()
			_ = conn.Close()
		}
		if err == nil {
			break
		}

		log.Printf("Waiting for %s: %s.", u, err)
		select {
		case <-time.After(cfg.delay):
		case <-ctx.Done():
			return
		}
	}

	readyc <- u
}
