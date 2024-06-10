//go:build generate

//go:generate mkdir -p .buildcache/bin
//go:generate -command GOINSTALL env "GOBIN=$PWD/.buildcache/bin" go install
//go:generate -command INSTALL-HADOLINT sh -c ".buildcache/bin/hadolint --version 2>/dev/null | grep -wq \"$DOLLAR{DOLLAR}{1}\" || curl -sSfL https://github.com/hadolint/hadolint/releases/download/v\"$DOLLAR{DOLLAR}{1}\"/hadolint-\"$(uname)\"-x86_64 --output ./.buildcache/bin/hadolint && chmod +x .buildcache/bin/hadolint" -sh
//go:generate -command INSTALL-SHELLCHECK sh -c ".buildcache/bin/shellcheck --version 2>/dev/null | grep -wq \"$DOLLAR{DOLLAR}{1}\" || curl -sSfL https://github.com/koalaman/shellcheck/releases/download/v\"$DOLLAR{DOLLAR}{1}\"/shellcheck-v\"$DOLLAR{DOLLAR}{1}\".\"$(uname)\".x86_64.tar.xz | tar xJf - -C .buildcache/bin --strip-components=1 shellcheck-v\"$DOLLAR{DOLLAR}{1}\"/shellcheck" -sh

package tools

//go:generate GOINSTALL github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2
//go:generate GOINSTALL github.com/mattn/goveralls@v0.0.12
//go:generate GOINSTALL github.com/tcnksm/ghr@v0.14.0
//go:generate GOINSTALL gotest.tools/gotestsum@v1.8.1
//go:generate INSTALL-HADOLINT 2.10.0
//go:generate INSTALL-SHELLCHECK 0.8.0
