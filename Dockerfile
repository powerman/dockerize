FROM golang:1.16.3-alpine3.13 AS builder
SHELL ["/bin/ash","-e","-o","pipefail","-x","-c"]

LABEL org.opencontainers.image.source="https://github.com/powerman/dockerize"

RUN apk add --no-cache openssl=~1.1.1k git=~2.30.2

COPY . /src
WORKDIR /src

RUN CGO_ENABLED=0 go install -ldflags "-X 'main.ver=$(git describe --match='v*' --exact-match)'"

FROM alpine:3.13

COPY --from=builder /go/bin/dockerize /usr/local/bin

ENTRYPOINT ["dockerize"]
CMD ["--help"]
