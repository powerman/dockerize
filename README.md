# dockerize

[![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/powerman/dockerize/test.yml?logo=github&label=build)](https://github.com/powerman/dockerize/actions/workflows/test.yml)
[![Coverage Status](https://raw.githubusercontent.com/powerman/dockerize/gh-badges/coverage.svg)](https://github.com/powerman/dockerize/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/powerman/dockerize)](https://goreportcard.com/report/github.com/powerman/dockerize)
[![Docker Automated Build](https://img.shields.io/docker/automated/powerman/dockerize.svg)](https://hub.docker.com/r/powerman/dockerize/tags)
[![Release](https://img.shields.io/github/v/release/powerman/dockerize.svg)](https://github.com/powerman/dockerize/releases/latest)

Utility to simplify running applications in docker containers.

**About this fork:**
This fork is supposed to become a community-maintained replacement for
[not maintained](https://github.com/powerman/dockerize/issues/19)
[original repo](https://github.com/jwilder/dockerize).
Everyone who has contributed to the project may become a collaborator -
just ask for it in PR comments after your PR has being merged.

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
  - [Docker Installation](#docker-installation)
  - [Go Install](#go-install)
- [Usage](#usage)
  - [Command-line Options](#command-line-options)
  - [Waiting for other dependencies](#waiting-for-other-dependencies)
  - [Timeout](#timeout)
  - [Delay before retrying](#delay-before-retrying)
  - [Use custom CA for SSL cert verification for https/amqps connections](#use-custom-ca-for-ssl-cert-verification-for-httpsamqps-connections)
  - [Skip SSL cert verification for https/amqps connections](#skip-ssl-cert-verification-for-httpsamqps-connections)
  - [Injecting env vars from INI file](#injecting-env-vars-from-ini-file)
  - [Customizing exit codes](#customizing-exit-codes)
- [Using Templates](#using-templates)
  - [jsonQuery](#jsonquery)

## Overview

dockerize is a utility to simplify running applications in docker containers.
It allows you to:

- generate application configuration files at container startup time
  from templates and container environment variables
- Tail multiple log files to stdout and/or stderr
- Wait for other services to be available using TCP, HTTP(S), unix
  before starting the main process.

The typical use case for dockerize is when you have an application
that has one or more configuration files
and you would like to control some of the values using environment variables.

For example, a Python application using Sqlalchemy
might not be able to use environment variables directly.
It may require that the database URL be read from a python settings file
with a variable named `SQLALCHEMY_DATABASE_URI`.
dockerize allows you to set an environment variable such as `DATABASE_URL`
and update the python file when the container starts.
In addition, it can also delay the starting of the python application
until the database container is running and listening on the TCP port.

Another use case is when the application logs to specific files on the filesystem
and not stdout or stderr.
This makes it difficult to troubleshoot the container using the `docker logs` command.
For example, nginx will log to `/var/log/nginx/access.log`
and `/var/log/nginx/error.log` by default.
While you can sometimes work around this,
it's tedious to find a solution for every application.
dockerize allows you to specify which logs files should be tailed
and where they should be sent.

See [A Simple Way To Dockerize Applications](http://jasonwilder.com/blog/2014/10/13/a-simple-way-to-dockerize-applications/)

## Installation

Dockerize available for these platforms:

| Platform      | Description                             |
| ------------- | --------------------------------------- |
| darwin-amd64  | Intel macOS                             |
| darwin-arm64  | Apple Silicon macOS                     |
| linux-386     | 32-bit Linux                            |
| linux-amd64   | 64-bit Linux                            |
| linux-arm64   | 64-bit ARM Linux (aarch64)              |
| linux-armv6   | 32-bit ARM Linux (Raspberry Pi 1, Zero) |
| linux-armv7   | 32-bit ARM Linux (Raspberry Pi 2, 3, 4) |
| linux-ppc64le | PowerPC 64-bit Linux                    |

To download it with most base images all you need is to install `curl` first:

```sh
### alpine:
apk add curl

### debian, ubuntu:
apt update && apt install -y curl
```

and then install
(replace `linux-amd64` with your platform from the table above):

```sh
curl -sfL https://github.com/powerman/dockerize/releases/download/v0.22.2/dockerize-v0.22.2-linux-amd64 | install /dev/stdin /usr/local/bin/dockerize
```

If `curl` is not available (e.g. busybox base image)
then you can use `wget`:

```sh
wget -O - https://github.com/powerman/dockerize/releases/download/v0.22.2/dockerize-v0.22.2-linux-amd64 | install /dev/stdin /usr/local/bin/dockerize
```

### Docker Installation

If you need to support multiple platforms in your Dockerfile,
there are two recommended approaches:

1. Use `powerman/dockerize` as base image -
   it's based on `alpine linux`,
   has dockerize pre-installed in `$PATH`,
   and available for all supported platforms:

   ```dockerfile
   FROM powerman/dockerize:0.22.2
   ...
   ENTRYPOINT dockerize ...
   ```

2. Copy dockerize from the base image into your image
   using multi-stage build:

   ```dockerfile
   FROM powerman/dockerize:0.22.2 AS dockerize
   FROM node:18-slim
   ...
   COPY --from=dockerize /usr/local/bin/dockerize /usr/local/bin/
   ...
   ENTRYPOINT ["dockerize", ...]
   ```

Both approaches will automatically use correct binary for your platform
without any manual platform detection.

### Go Install

If you have Go installed, you can install dockerize using `go install`:

```sh
go install github.com/powerman/dockerize@latest
```

## Usage

dockerize works by wrapping the call to your application
using the `ENTRYPOINT` or `CMD` directives.

This would generate `/etc/nginx/nginx.conf` from the template located at `/etc/nginx/nginx.tmpl`
and send `/var/log/nginx/access.log` to `STDOUT`
and `/var/log/nginx/error.log` to `STDERR` after running `nginx`,
only after waiting for the `web` host to respond on `tcp 8000`:

```dockerfile
CMD dockerize \
  -template /etc/nginx/nginx.tmpl:/etc/nginx/nginx.conf \
  -stdout /var/log/nginx/access.log \
  -stderr /var/log/nginx/error.log \
  -wait tcp://web:8000 \
  nginx
```

### Command-line Options

You can specify multiple templates
by passing using `-template` multiple times:

```sh
dockerize -template template1.tmpl:file1.cfg -template template2.tmpl:file3
```

Templates can be generated to `STDOUT`
by not specifying a dest:

```sh
dockerize -template template1.tmpl
```

Template may also be a directory.
In this case all files within this directory are recursively processed as template
and stored with the same name in the destination directory.
If the destination directory is omitted,
the output is sent to `STDOUT`.
The files in the source directory are processed in sorted order
(as returned by `os.ReadDir`).

```sh
dockerize -template src_dir:dest_dir
```

If the destination file already exists,
dockerize will overwrite it.
The `-no-overwrite` flag overrides this behaviour.

```sh
dockerize -no-overwrite -template template1.tmpl:file
```

You can tail multiple files to `STDOUT` and `STDERR`
by passing the options multiple times.
(These options can't be combined with `-exec`.)

```sh
dockerize -stdout info.log -stdout perf.log
```

If your file uses `{{` and `}}` as part of it's syntax,
you can change the template escape characters using the `-delims`.

```sh
dockerize -delims "<%:%>" -template template1.tmpl
```

You can require all environment variables mentioned in template exists
with `-template-strict`:

```sh
dockerize -template-strict -template template1.tmpl
```

HTTP headers can be specified for http/https protocols.
If header is specified as a file path
then file must contain single string with `Header: value`.

```sh
dockerize -wait http://web:80 \
    -wait-http-header "Authorization:Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ=="
```

Required HTTP status codes can be specified,
otherwise any 2xx status will be accepted.

```sh
dockerize -wait http://web:80 \
    -wait-http-status-code 302 \
    -wait-http-status-code 200
```

HTTP redirects can be ignored:

```sh
dockerize -wait http://web:80 -wait-http-skip-redirect
```

Dockerize process can be replaced with given command:

```sh
dockerize -exec some-command args...
```

### Waiting for other dependencies

It is common when using tools like [Docker Compose](https://docs.docker.com/compose/)
to depend on services in other linked containers,
however oftentimes relying on [links](https://docs.docker.com/compose/compose-file/#links) is not enough -
whilst the container itself may have _started_,
the _service(s)_ within it may not yet be ready -
resulting in shell script hacks to work around race conditions.

Dockerize gives you the ability to wait for services on a specified protocol
(`file`, `tcp`, `tcp4`, `tcp6`, `http`, `https`, `amqp`, `amqps` and `unix`)
before starting your application:

```sh
dockerize -wait tcp://db:5432 \
    -wait http://web:80 \
    -wait file:///tmp/generated-file
```

Multiple URLs can also be specified with `-wait-list` flag,
that accept a space-separated list of URLs.
The behaviour is equivalent to use multiple `-wait` flags.
The two flags can be combined.

This command is equivalent to the one above:

```sh
dockerize -wait-list "tcp://db:5432 http://web:80 file:///tmp/generated-file"
```

### Timeout

You can optionally specify how long to wait for the services to become available
by using the `-timeout #` argument (Default: 10 seconds).
If the timeout is reached and the service is still not available,
the process exits with status code 123.

```sh
dockerize -wait tcp://db:5432 \
    -wait http://web:80 \
    -timeout 10s
```

See [this issue](https://github.com/docker/compose/issues/374#issuecomment-126312313)
for a deeper discussion,
and why support isn't and won't be available in the Docker ecosystem itself.

### Delay before retrying

You can optionally specify how long to wait after a failed `-wait` check
by using the `-wait-retry-interval #` argument (Default: 1 second).

Waiting for 5 seconds before checking again
of a currently unavailable service:

```sh
dockerize -wait tcp://db:5432 -wait-retry-interval 5s
```

### Use custom CA for SSL cert verification for https/amqps connections

```sh
dockerize -cacert /path/to/ca.pem -wait https://web:80
```

### Skip SSL cert verification for https/amqps connections

```sh
dockerize -skip-tls-verify -wait https://web:80
```

### Injecting env vars from INI file

You can load defaults for missing env vars from INI file.
Multiline flag allows parsing multiline INI entries.
File with header must contain single string with `Header: value`.

```sh
dockerize -env /path/to/file.ini \
    -env-section SectionName \
    -multiline …
dockerize -env http://localhost:80/file.ini \
    -env-header "Header: value" \
    -env-header /path/to/file/with/header …
```

### Customizing exit codes

By default, dockerize exits with code 123 when encountering errors.
You can customize this using the `-exit-code` flag:

```sh
dockerize -exit-code 42 -wait tcp://db:5432 app
```

## Using Templates

Templates use Golang [text/template](http://golang.org/pkg/text/template/).
You can access environment variables within a template with `.Env`.

```gotmpl
{{ .Env.PATH }} is my path
```

In template you can use a lot of
[functions provided by Sprig](http://masterminds.github.io/sprig/)
plus a few built in functions as well:

- `exists $path` - Determines if a file path exists or not.
  `{{ if exists "/etc/default/myapp" }}`
- `parseUrl $url` - Parses a URL into it's
  [protocol, scheme, host, etc. parts](https://golang.org/pkg/net/url/#URL).
  Alias for [`url.Parse`](https://golang.org/pkg/net/url/#Parse)
  `{{ (parseUrl "https://example.com/path").Host }}`
- `isTrue $value` - Parses a string $value to a boolean value.
  `{{ if isTrue .Env.ENABLED }}`
- `jsonQuery $json $query` - Returns the result of a selection query
  against a json document.
- `readFile $fileName` - Returns the content of the named file
  or empty string if file not exists.
  `{{ readFile "/etc/hostname" }}`

**WARNING! Incompatibility with [original dockerize v0.6.1](https://github.com/jwilder/dockerize)!**
These template functions was changed because of adding Sprig functions,
so carefully review your templates before upgrading:

- `default` - order of params has changed.
- `contains` - now it works on string instead of map, use `hasKey` instead.
- `split` - now it split into map instead of list, use `splitList` instead.
- `replace` - order and amount of params has changed.
- `loop` - removed, use `untilStep` instead.

### jsonQuery

Objects and fields are accessed by name.
Array elements are accessed by index in square brackets (e.g. `[1]`).
Nested elements are separated by dots (`.`).

**Examples:**

With the following JSON in `.Env.SERVICES`

```json
{
  "services": [
    {
      "name": "service1",
      "port": 8000
    },
    {
      "name": "service2",
      "port": 9000
    }
  ]
}
```

the template expression `jsonQuery .Env.SERVICES "services.[1].port"`
returns `9000`.
