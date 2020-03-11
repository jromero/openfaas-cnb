# Samples

Samples for how this Cloud Native Buildpack is integrated with OpenFaaS.

### Prerequisites

- [Docker](https://docs.docker.com/install/)
- [Kind](https://github.com/kubernetes-sigs/kind#installation-and-usage)
- [faas-cli](https://github.com/openfaas/faas-cli#get-started-install-the-cli)
- [pack](https://buildpacks.io/docs/install-pack/)

### Environment

#### Start up

```shell script
./scripts/local.startup.sh
```

This command does the following:

- Start up a docker registry
- Starts a k8s cluster
- Installs OpenFaaS

#### Teardown

```shell script
./scripts/local.teardown.sh
```

### Build

Build and publish application to local registry.

```shell script
pack build localhost:5001/openfaas-cnb/streaming:latest \
  --path streaming/ \
  --builder heroku/buildpacks:18 \
  --buildpack from=builder \
  --buildpack jar013/openfaas-cnb:latest \
  --publish
```

### Run

```shell script
faas up -f stack.yml
```

### Troubleshooting

Some helpful commands:

```shell script
# list function pods 
kubectl -n openfaas-fn get pods

# details of function pods
kubectl -n openfaas-fn describe pods
```