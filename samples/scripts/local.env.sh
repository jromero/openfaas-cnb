#!/usr/bin/env bash

export OPENFAAS_CLUSTER=openfaas-cbn-samples
export OPENFAAS_CONTEXT=kind-${OPENFAAS_CLUSTER}
export OPENFAAS_PORT=8989
export OPENFAAS_URL=http://127.0.0.1:${OPENFAAS_PORT}
export OPENFAAS_REG_NAME=${OPENFAAS_CLUSTER}-registry
export OPENFAAS_REG_PORT=5001

export | grep OPENFAAS

knd() {
  kind --name "${OPENFAAS_CLUSTER}" "$@"
}

kc() {
  kubectl --context "${OPENFAAS_CONTEXT}" "$@"
}

hlm() {
  helm --kube-context="${OPENFAAS_CONTEXT}" "$@"
}