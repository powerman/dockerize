package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
	"strings"
)

var (
	errNameValueMultiline   = errors.New("name:value must be a single line")
	errNameValueRequired    = errors.New("must be a name:value")
	errStatusCodeOutOfRange = errors.New("status code must be between 100 and 599")
	errLeftRightRequired    = errors.New("must be a left:right")
)

type stringsFlag []string

func (f *stringsFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func (f stringsFlag) String() string {
	return strings.Join(f, ",")
}

type urlsFlag []*url.URL

func (f *urlsFlag) Set(value string) error {
	u, err := url.Parse(value)
	if err != nil {
		return err
	}
	*f = append(*f, u)
	return nil
}

func (f urlsFlag) String() string {
	urls := make([]string, len(f))
	for i := range f {
		urls[i] = f[i].String()
	}
	return strings.Join(urls, " ")
}

type httpHeader struct {
	name  string
	value string
}

type httpHeadersFlag []httpHeader

func (f *httpHeadersFlag) Set(value string) error {
	buf, err := ioutil.ReadFile(value) //nolint:gosec // File inclusion via variable.
	if err == nil {
		value = string(buf)
	} else if !os.IsNotExist(err) {
		return err
	}
	value = strings.TrimSpace(value)
	if strings.ContainsAny(value, "\r\n") {
		return errNameValueMultiline
	} else if strings.Count(value, ":") == 0 {
		return errNameValueRequired
	}
	nv := strings.SplitN(value, ":", 2)
	for i := range nv {
		nv[i] = strings.TrimSpace(nv[i])
		if nv[i] == "" {
			return errNameValueRequired
		}
	}
	*f = append(*f, httpHeader{name: nv[0], value: nv[1]})
	return nil
}

func (f httpHeadersFlag) String() string {
	hs := make([]string, len(f))
	for i := range f {
		hs[i] = f[i].name + ":" + f[i].value
	}
	return strings.Join(hs, ", ")
}

type statusCodesFlag []int

func (f *statusCodesFlag) Set(value string) error {
	i, err := strconv.Atoi(value)
	if err != nil {
		return err
	}
	if i < 100 || 599 < i {
		return errStatusCodeOutOfRange
	}
	*f = append(*f, i)
	return nil
}

func (f statusCodesFlag) String() string {
	ns := make([]string, len(f))
	for i := range f {
		ns[i] = strconv.Itoa(f[i])
	}
	return strings.Join(ns, ", ")
}

type delimsFlag [2]string

func (f *delimsFlag) Set(value string) error {
	delims := strings.Split(value, ":")
	if len(delims) != 2 || len(strings.Fields(delims[0])) != 1 || len(strings.Fields(delims[1])) != 1 {
		return errLeftRightRequired
	}
	(*f)[0] = delims[0]
	(*f)[1] = delims[1]
	return nil
}

func (f delimsFlag) String() string {
	return f[0] + ":" + f[1]
}

// fatalFlagValue report invalid flag values in same way as flag.Parse().
func fatalFlagValue(msg, name string, val interface{}) {
	fmt.Fprintf(os.Stderr, "invalid value %q for flag -%s: %s\n", val, name, msg)
	flag.Usage()
	os.Exit(exitCodeUsage)
}
