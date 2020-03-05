# openfaas-cnb ![Build](https://github.com/jromero/openfaas-cnb/workflows/Build/badge.svg)

Clound Native Buildpack for [OpenFaaS](https://www.openfaas.com/)

## Getting started

### Usage

This buildpack is currently intended to be used with the heroku builder [heroku/pack:18](https://github.com/heroku/pack-images) via [`pack`](https://github.com/buildpacks/pack).

#### Configuration

An _optional_ `watchdog.toml` configuration file may be present at the application root:

```toml
[watchdog]
# The watchdog version to use.
# See https://github.com/openfaas-incubator/of-watchdog/releases
# (default: 0.7.6)
version = "0.7.6"

# The cloud native buildpack process type to run.
# See `pack inspect-image <built-app-image>`
# (default: web)
process_type = "web"

# Key-value pairs that will be passed onto image as environment variables.
# See https://github.com/openfaas-incubator/of-watchdog#configuration
[watchdog.env]
```

#### Build your app

```shell script
pack build my-app \
  --builder heroku/buildpacks:18 \
  --buildpack from=builder \
  --buildpack jar03/openfass-cnb:latest \
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
make package-tgz
```

Creates a distributable buildpack image. 

```shell script
make package-image
```

#### Troubleshooting

```shell script
pack build ... -e BP_DEBUG=true
```