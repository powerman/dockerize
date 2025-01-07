# syntax=docker/dockerfile:1

# Go version is also in .github/workflows/CI&CD.yml.
FROM golang:1.23.2-alpine3.20 AS builder
SHELL ["/bin/ash","-e","-o","pipefail","-x","-c"]

LABEL org.opencontainers.image.source="https://github.com/powerman/dockerize"

# hadolint ignore=DL3019
RUN apk update; \
    apk add openssl=~3 git=~2; \
    apk add upx=~4 || :; \
    rm -f /var/cache/apk/*

COPY . /src
WORKDIR /src

RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 go install -ldflags "-s -w -X 'main.ver=$(git describe --match='v*' --exact-match)'" && \
    ! which upx >/dev/null || upx /go/bin/dockerize && \
    dockerize --version

FROM alpine:3.21.1

COPY --from=builder /go/bin/dockerize /usr/local/bin

ENTRYPOINT ["dockerize"]
CMD ["--help"]
