# Go version is also in .github/workflows/CI&CD.yml.
FROM golang:1.17.2-alpine3.14 AS builder
SHELL ["/bin/ash","-e","-o","pipefail","-x","-c"]

LABEL org.opencontainers.image.source="https://github.com/powerman/dockerize"

RUN apk add --no-cache openssl=~1.1.1l git=~2.32.0

COPY . /src
WORKDIR /src

RUN CGO_ENABLED=0 go install -ldflags "-X 'main.ver=$(git describe --match='v*' --exact-match)'"

FROM alpine:3.14

COPY --from=builder /go/bin/dockerize /usr/local/bin

ENTRYPOINT ["dockerize"]
CMD ["--help"]
