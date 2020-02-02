# openfaas-cnb ![Build](https://github.com/jromero/openfaas-cnb/workflows/Build/badge.svg)

Clound Native Buildpack for [OpenFaaS](https://www.openfaas.com/)

## Getting started

### Usage

This buildpack is currently intended to be used with the heroku builder [heroku/pack:18](https://github.com/heroku/pack-images) via [`pack`](https://github.com/buildpacks/pack).

#### Configuration

A `watchdog.toml` must be present at the application root.

```toml
[watchdog]
# The watchdog version to use.
# See https://github.com/openfaas-incubator/of-watchdog/releases
version = "0.7.6"

# Key-value pairs that will be passed onto image as environment variables.
# See https://github.com/openfaas-incubator/of-watchdog#configuration
[watchdog.env]
function_process = "./app.sh"
```

#### Build your app

```shell script
pack build my-app \
  --builder heroku/buildpacks:18 \
  --buildpack from=builder \
  --buildpack https://github.com/jromero/openfaas-cnb/releases/download/0.0.2/openfaas-cnb-0.0.2.tgz \
  --path .
```


### Building

```shell script
make build
```

### Testing (End-to-End)

Run end-to-end test using [`pack`](https://github.com/buildpacks/pack).

```shell script
make test-e2e
```

### Packaging

Creates a portable `.tgz` format of this buildpack. 

```shell script
make package
```
