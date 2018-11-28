package main

import (
	"crypto/tls"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
)

type iniConfig struct { // nolint:maligned
	source        string // URL or file path
	multiline     bool
	section       string
	headers       httpHeadersFlag
	skipTLSVerify bool
}

func getINI(cfg iniConfig) (data []byte, err error) {
	// See if envFlag parses like an absolute URL, if so use http, otherwise treat as filename
	url, urlERR := url.ParseRequestURI(cfg.source)
	if urlERR == nil && url.IsAbs() {
		var resp *http.Response
		var req *http.Request
		var client *http.Client
		// Define redirect handler to disallow redirects
		var redir = func(req *http.Request, via []*http.Request) error {
			return errors.New("redirects disallowed")
		}

		transport := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.skipTLSVerify},
		}
		client = &http.Client{Transport: transport, CheckRedirect: redir}
		req, err = http.NewRequest("GET", cfg.source, nil)
		if err != nil {
			// Weird problem with declaring client, bail
			return data, err
		}
		// Handle headers for request - are they headers or filepaths?
		for _, h := range cfg.headers {
			req.Header.Add(h.name, h.value)
		}
		resp, err = client.Do(req)
		if err == nil && resp.StatusCode == 200 {
			defer resp.Body.Close()
			data, err = ioutil.ReadAll(resp.Body)
		} else if err == nil { // Request completed with unexpected HTTP status code, bail
			err = errors.New(resp.Status)
			return data, err
		}
	} else {
		data, err = ioutil.ReadFile(cfg.source)
	}
	return data, err
}
