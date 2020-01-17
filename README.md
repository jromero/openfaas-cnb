> This buildpack currently requires additional functionality to be provided by pack. Specifically [PR#450](https://github.com/buildpacks/pack/pull/450).

# openfaas-cnb

Clound Native Buildpack for [OpenFaaS](https://www.openfaas.com/)

## Getting started

### Usage

This buildpack is currently intended to be used with the heroku builder [heroku/pack:18](https://github.com/heroku/pack-images) via [`pack`](https://github.com/buildpacks/pack).

#### Configuration

A `watchdog.toml` must be present at the application root.

```toml
[watchdog]
# The watchdog version to use.
# See https://github.com/openfaas/faas/releases 
version = "0.9.14"

# Key-value pairs that will be passed onto image as environment variables.
# See https://github.com/openfaas/faas/tree/master/watchdog#delve-deeper
[watchdog.env]
fprocess = "./app.sh"
cgi_headers = "true"
```

#### Build your app

```shell script
pack build my-app \
  --builder heroku/buildpacks:18 \
  --buildpack from=builder \
  --buildpack https://github.com/jromero/openfaas-cnb/releases/download/0.0.1/openfaas-cnb-0.0.1.tgz \
  --path .
```

### Testing

```shell script
make test
```

### Packaging

Creates a portable `.tgz` format of this buildpack. 

```shell script
make package
```
