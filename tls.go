package main

import (
	"crypto/x509"
	"errors"
	"fmt"
	"os"
)

var errPEM = errors.New("unable to load PEM certs")

// LoadCACert returns a new CertPool with certificates loaded from given path.
func LoadCACert(path string) (*x509.CertPool, error) {
	if path == "" {
		return nil, nil //nolint:nilnil // TODO.
	}
	ca, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}
	caCert, err := os.ReadFile(path) //nolint:gosec // False positive.
	if err == nil && !ca.AppendCertsFromPEM(caCert) {
		err = fmt.Errorf("%w: %q", errPEM, path)
	}
	if err != nil {
		return nil, err
	}
	return ca, nil
}
