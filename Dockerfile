# Go version is also in .github/workflows/CI&CD.yml.
FROM golang:1.20.1-alpine3.17 AS builder
SHELL ["/bin/ash","-e","-o","pipefail","-x","-c"]

LABEL org.opencontainers.image.source="https://github.com/powerman/dockerize"

RUN apk add --no-cache openssl=~3 git=~2

COPY . /src
WORKDIR /src

RUN CGO_ENABLED=0 go install -ldflags "-s -X 'main.ver=$(git describe --match='v*' --exact-match)'"

FROM alpine:3.17.2

COPY --from=builder /go/bin/dockerize /usr/local/bin

ENTRYPOINT ["dockerize"]
CMD ["--help"]
