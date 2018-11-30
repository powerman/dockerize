FROM golang:1.11-alpine3.7 AS binary
RUN apk -U add openssl git

ADD . /src
WORKDIR /src

RUN CGO_ENABLED=0 go install -ldflags "-X 'main.ver=$(git describe --match='v*' --exact-match)'"

FROM alpine:3.7

COPY --from=binary /go/bin/dockerize /usr/local/bin

ENTRYPOINT ["dockerize"]
CMD ["--help"]
