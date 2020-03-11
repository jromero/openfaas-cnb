#!/usr/bin/env bash
set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
pushd "${DIR}"

source local.env.sh

echo "> Stoping registry..."
docker stop "${OPENFAAS_REG_NAME}" || true
docker rm "${OPENFAAS_REG_NAME}" || true

echo "> Deleting cluster..."
if [[ $(kind get clusters | grep -c "${OPENFAAS_CLUSTER}") -ge 1 ]]; then
  knd delete cluster
fi