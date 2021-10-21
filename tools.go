//go:build tools || generate
// +build tools generate

//go:generate sh -c "GOBIN=$PWD/.gobincache go install $(sed -n 's/.*_ \"\\(.*\\)\".*/\\1/p' <$GOFILE)"

package main

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/mattn/goveralls"
	_ "github.com/tcnksm/ghr"
	_ "gotest.tools/gotestsum"
)
