# dockerize [![GitHub release](https://img.shields.io/github/release/powerman/dockerize.svg)](https://github.com/powerman/dockerize/releases/latest) [![CI Build Status](https://circleci.com/gh/powerman/dockerize.svg?style=svg)](https://circleci.com/gh/powerman/dockerize) [![Docker Build Status](https://img.shields.io/docker/build/powerman/dockerize.svg)](https://hub.docker.com/r/powerman/dockerize/) [![Go Report Card](https://goreportcard.com/badge/github.com/powerman/dockerize)](https://goreportcard.com/report/github.com/powerman/dockerize) [![Coverage Status](https://coveralls.io/repos/github/powerman/dockerize/badge.svg?branch=master)](https://coveralls.io/github/powerman/dockerize?branch=master)

Utility to simplify running applications in docker containers.

**About this fork:** This fork is supposed to become a community-maintained replacement for
[not maintained](https://github.com/powerman/dockerize/issues/19)
[original repo](https://github.com/jwilder/dockerize). Everyone who has
contributed to the project may become a collaborator - just ask for it
in PR comments after your PR has being merged.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**

- [Overview](#overview)
- [Installation](#installation)
  - [Docker Base Image](#docker-base-image)
- [Usage](#usage)
  - [Command-line Options](#command-line-options)
  - [Waiting for other dependencies](#waiting-for-other-dependencies)
  - [Timeout](#timeout)
  - [Use custom CA for SSL cert verification for https/amqps connections](#use-custom-ca-for-ssl-cert-verification-for-httpsamqps-connections)
  - [Skip SSL cert verification for https/amqps connections](#skip-ssl-cert-verification-for-httpsamqps-connections)
  - [Injecting env vars from INI file](#injecting-env-vars-from-ini-file)
- [Using Templates](#using-templates)
  - [jsonQuery](#jsonquery)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Overview

dockerize is a utility to simplify running applications in docker containers.  It allows you to:
* generate application configuration files at container startup time from templates and container environment variables
* Tail multiple log files to stdout and/or stderr
* Wait for other services to be available using TCP, HTTP(S), unix before starting the main process.

The typical use case for dockerize is when you have an application that has one or more configuration files and you would like to control some of the values using environment variables.

For example, a Python application using Sqlalchemy might not be able to use environment variables directly.
It may require that the database URL be read from a python settings file with a variable named
`SQLALCHEMY_DATABASE_URI`.  dockerize allows you to set an environment variable such as
`DATABASE_URL` and update the python file when the container starts.
In addition, it can also delay the starting of the python application until the database container is running and listening on the TCP port.

Another use case is when the application logs to specific files on the filesystem and not stdout
or stderr. This makes it difficult to troubleshoot the container using the `docker logs` command.
For example, nginx will log to `/var/log/nginx/access.log` and
`/var/log/nginx/error.log` by default. While you can sometimes work around this, it's tedious to find a solution for every application. dockerize allows you to specify which logs files should be tailed and where they should be sent.

See [A Simple Way To Dockerize Applications](http://jasonwilder.com/blog/2014/10/13/a-simple-way-to-dockerize-applications/)


## Installation

Dockerize is a statically compiled binary, so it should work with any base image.

To download it with most base images all you need is to install `curl` first:

```sh
### alpine:
apk add curl

### debian, ubuntu:
apt update && apt install -y curl
```

and then either install the latest version:

```sh
curl -sfL $(curl -s https://api.github.com/repos/powerman/dockerize/releases/latest | grep -i /dockerize-$(uname -s)-$(uname -m)\" | cut -d\" -f4) | install /dev/stdin /usr/local/bin/dockerize
```

or specific version:

```sh
curl -sfL https://github.com/powerman/dockerize/releases/download/v0.11.5/dockerize-`uname -s`-`uname -m` | install /dev/stdin /usr/local/bin/dockerize
```

If `curl` is not available (e.g. busybox base image) then you can use `wget`:

```
### busybox: latest version
wget -O - $(wget -O - https://api.github.com/repos/powerman/dockerize/releases/latest | grep -i /dockerize-$(uname -s)-$(uname -m)\" | cut -d\" -f4) | install /dev/stdin /usr/local/bin/dockerize

### busybox: specific version
wget -O - https://github.com/powerman/dockerize/releases/download/v0.11.5/dockerize-`uname -s`-`uname -m` | install /dev/stdin /usr/local/bin/dockerize
```

PGP public key for verifying signed binaries: https://powerman.name/about/Powerman.asc

```
curl -sfL https://powerman.name/about/Powerman.asc | gpg --import
curl -sfL https://github.com/powerman/dockerize/releases/download/v0.11.5/dockerize-`uname -s`-`uname -m`.asc >dockerize.asc
gpg --verify dockerize.asc /usr/local/bin/dockerize
```

### Docker Base Image

The `powerman/dockerize` image is a base image based on `alpine linux`.  `dockerize` is installed in the `$PATH` and can be used directly.

```
FROM powerman/dockerize
...
ENTRYPOINT dockerize ...
```

## Usage

dockerize works by wrapping the call to your application using the `ENTRYPOINT` or `CMD` directives.

This would generate `/etc/nginx/nginx.conf` from the template located at `/etc/nginx/nginx.tmpl` and
send `/var/log/nginx/access.log` to `STDOUT` and `/var/log/nginx/error.log` to `STDERR` after running
`nginx`, only after waiting for the `web` host to respond on `tcp 8000`:

``` Dockerfile
CMD dockerize -template /etc/nginx/nginx.tmpl:/etc/nginx/nginx.conf -stdout /var/log/nginx/access.log -stderr /var/log/nginx/error.log -wait tcp://web:8000 nginx
```

### Command-line Options

You can specify multiple templates by passing using `-template` multiple times:

```
$ dockerize -template template1.tmpl:file1.cfg -template template2.tmpl:file3

```

Templates can be generated to `STDOUT` by not specifying a dest:

```
$ dockerize -template template1.tmpl

```

Template may also be a directory. In this case all files within this directory are recursively processed as template and stored with the same name in the destination directory.
If the destination directory is omitted, the output is sent to `STDOUT`. The files in the source directory are processed in sorted order (as returned by `ioutil.ReadDir`).

```
$ dockerize -template src_dir:dest_dir

```

If the destination file already exists, dockerize will overwrite it. The -no-overwrite flag overrides this behaviour.

```
$ dockerize -no-overwrite -template template1.tmpl:file
```

You can tail multiple files to `STDOUT` and `STDERR` by passing the options multiple times.

```
$ dockerize -stdout info.log -stdout perf.log

```

If your file uses `{{` and `}}` as part of it's syntax, you can change the template escape characters using the `-delims`.

```
$ dockerize -delims "<%:%>" -template template1.tmpl
```

You can require all environment variables mentioned in template exists
with `-template-strict`:

```
$ dockerize -template-strict -template template1.tmpl
```

HTTP headers can be specified for http/https protocols.
If header is specified as a file path then file must contain single string with `Header: value`.

```
$ dockerize -wait http://web:80 -wait-http-header "Authorization:Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ=="
```

Required HTTP status codes can be specified, otherwise any 2xx status will
be accepted.

```
$ dockerize -wait http://web:80 -wait-http-status-code 302 -wait-http-status-code 200
```

HTTP redirects can be ignored:

```
$ dockerize -wait http://web:80 -wait-http-skip-redirect
```

### Waiting for other dependencies

It is common when using tools like [Docker Compose](https://docs.docker.com/compose/) to depend on services in other linked containers, however oftentimes relying on [links](https://docs.docker.com/compose/compose-file/#links) is not enough - whilst the container itself may have _started_, the _service(s)_ within it may not yet be ready - resulting in shell script hacks to work around race conditions.

Dockerize gives you the ability to wait for services on a specified protocol (`file`, `tcp`, `tcp4`, `tcp6`, `http`, `https`, `amqp`, `amqps` and `unix`) before starting your application:

```
$ dockerize -wait tcp://db:5432 -wait http://web:80 -wait file:///tmp/generated-file
```

### Timeout

You can optionally specify how long to wait for the services to become available by using the `-timeout #` argument (Default: 10 seconds).  If the timeout is reached and the service is still not available, the process exits with status code 123.

```
$ dockerize -wait tcp://db:5432 -wait http://web:80 -timeout 10s
```

See [this issue](https://github.com/docker/compose/issues/374#issuecomment-126312313) for a deeper discussion, and why support isn't and won't be available in the Docker ecosystem itself.

### Use custom CA for SSL cert verification for https/amqps connections

```
$ dockerize -cacert /path/to/ca.pem -wait https://web:80
```

### Skip SSL cert verification for https/amqps connections

```
$ dockerize -skip-tls-verify -wait https://web:80
```

### Injecting env vars from INI file

You can load defaults for missing env vars from INI file.
Multiline flag allows parsing multiline INI entries.
File with header must contain single string with `Header: value`.

```
$ dockerize -env /path/to/file.ini -env-section SectionName -multiline …
$ dockerize -env http://localhost:80/file.ini \
    -env-header "Header: value" -env-header /path/to/file/with/header …
```

## Using Templates

Templates use Golang [text/template](http://golang.org/pkg/text/template/). You can access environment
variables within a template with `.Env`.

```
{{ .Env.PATH }} is my path
```

In template you can use a lot of [functions provided by
Sprig](http://masterminds.github.io/sprig/) plus a few built in functions as well:

  * `exists $path` - Determines if a file path exists or not. `{{ if exists "/etc/default/myapp" }}`
  * `parseUrl $url` - Parses a URL into it's [protocol, scheme, host, etc. parts](https://golang.org/pkg/net/url/#URL). Alias for [`url.Parse`](https://golang.org/pkg/net/url/#Parse)
  * `isTrue $value` - Parses a string $value to a boolean value. `{{ if isTrue .Env.ENABLED }}`
  * `jsonQuery $json $query` - Returns the result of a selection query against a json document.
  * `readFile $fileName` - Returns the content of the named file or empty string if file not exists.

**WARNING! Incompatibility with [original dockerize
v0.6.1](https://github.com/jwilder/dockerize)!** These template functions
was changed because of adding Sprig functions, so carefully review your
templates before upgrading:

* `default` - order of params has changed.
* `contains` - now it works on string instead of map, use `hasKey` instead.
* `split` - now it split into map instead of list, use `splitList` instead.
* `replace` - order and amount of params has changed.
* `loop` - removed, use `untilStep` instead.

### jsonQuery

Objects and fields are accessed by name. Array elements are accessed by index in square brackets (e.g. `[1]`). Nested elements are separated by dots (`.`).

**Examples:**

With the following JSON in `.Env.SERVICES`

```
{
  "services": [
    {
      "name": "service1",
      "port": 8000,
    },{
      "name": "service2",
      "port": 9000,
    }
  ]
}
```

the template expression `jsonQuery .Env.SERVICES "services.[1].port"` returns `9000`.
