package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	ini "gopkg.in/ini.v1"
)

var (
	errRedirectsDisallowed = errors.New("redirects disallowed")
	errBadStatusCode       = errors.New("bad HTTP status")
)

type iniConfig struct {
	source        string // URL or file path
	options       ini.LoadOptions
	section       string
	headers       httpHeadersFlag
	skipTLSVerify bool
	ca            *x509.CertPool
}

func loadINISection(cfg iniConfig) (map[string]string, error) {
	if cfg.source == "" {
		return nil, nil
	}

	var data []byte
	u, err := url.Parse(cfg.source)
	if err == nil && u.IsAbs() {
		data, err = fetchINI(cfg)
	} else {
		data, err = ioutil.ReadFile(cfg.source)
	}
	if err != nil {
		return nil, err
	}

	file, err := ini.LoadSources(cfg.options, data)
	if err != nil {
		return nil, err
	}
	return file.Section(cfg.section).KeysHash(), nil
}

func fetchINI(cfg iniConfig) (data []byte, err error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: cfg.skipTLSVerify, //nolint:gosec // TLS InsecureSkipVerify may be true.
				RootCAs:            cfg.ca,
			},
		},
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return errRedirectsDisallowed
		},
	}

	req, err := http.NewRequestWithContext(context.Background(), "GET", cfg.source, nil)
	if err != nil {
		return nil, err
	}
	for _, h := range cfg.headers { //nolint:gocritic // Premature optimization.
		req.Header.Add(h.name, h.value)
	}

	resp, err := client.Do(req) //nolint:bodyclose // False positive.
	if err != nil {
		return nil, err
	}
	defer warnIfFail(resp.Body.Close)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", errBadStatusCode, resp.StatusCode)
	}
	return ioutil.ReadAll(resp.Body)
}
