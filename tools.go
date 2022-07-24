//go:build generate

//go:generate mkdir -p .gobincache
//go:generate -command GOINSTALL env "GOBIN=$PWD/.gobincache" go install
//go:generate -command INSTALL-HADOLINT sh -c ".gobincache/hadolint --version 2>/dev/null | grep -wq \"$DOLLAR{DOLLAR}{1}\" || curl -sSfL https://github.com/hadolint/hadolint/releases/download/v\"$DOLLAR{DOLLAR}{1}\"/hadolint-\"$(uname)\"-x86_64 --output ./.gobincache/hadolint && chmod +x .gobincache/hadolint" -sh
//go:generate -command INSTALL-SHELLCHECK sh -c ".gobincache/shellcheck --version 2>/dev/null | grep -wq \"$DOLLAR{DOLLAR}{1}\" || curl -sSfL https://github.com/koalaman/shellcheck/releases/download/v\"$DOLLAR{DOLLAR}{1}\"/shellcheck-v\"$DOLLAR{DOLLAR}{1}\".\"$(uname)\".x86_64.tar.xz | tar xJf - -C .gobincache --strip-components=1 shellcheck-v\"$DOLLAR{DOLLAR}{1}\"/shellcheck" -sh

package main

//go:generate GOINSTALL github.com/golangci/golangci-lint/cmd/golangci-lint@v1.45.2
//go:generate GOINSTALL github.com/mattn/goveralls@v0.0.11
//go:generate GOINSTALL github.com/tcnksm/ghr@v0.14.0
//go:generate GOINSTALL gotest.tools/gotestsum@v1.8.1
//go:generate INSTALL-HADOLINT 2.10.0
//go:generate INSTALL-SHELLCHECK 0.8.0
