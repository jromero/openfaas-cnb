#!/usr/bin/env bash
set -e

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
pushd "${DIR}"

source local.env.sh

echo "> Starting registry..."
REG_RUNNING="$(docker inspect -f '{{.State.Running}}' "${OPENFAAS_REG_NAME}" 2>/dev/null || true)"
if [ "${REG_RUNNING}" != 'true' ]; then
  docker run -d -p "${OPENFAAS_REG_PORT}:5000" --name "${OPENFAAS_REG_NAME}" registry:2
fi
REG_IP="$(docker inspect -f '{{.NetworkSettings.IPAddress}}' "${OPENFAAS_REG_NAME}")"

echo "> Create cluster..."
if [[ $(kind get clusters | grep -c "^${OPENFAAS_CLUSTER}") -eq 0 ]]; then
  cat <<EOF | knd create cluster --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:5000"]
    endpoint = ["http://${REG_IP}:5000"]
EOF

else
  echo "Cluster already running!"
fi

echo "> Creating OpenFaaS namespaces..."
kc apply -f https://raw.githubusercontent.com/openfaas/faas-netes/master/namespaces.yml

echo "> Configuring Helm..."
hlm repo add openfaas https://openfaas.github.io/faas-netes/
hlm repo update

echo "> Install OpenFaas..."
hlm upgrade openfaas --install openfaas/openfaas \
  --namespace openfaas \
  --set functionNamespace=openfaas-fn \
  --set basic_auth=false \
  --set operator.create=true

echo "> Waiting for OpenFaas pods to be ready..."
sleep 10 # wait for initial registration
kc wait --timeout=90s --for=condition=Ready pods -n openfaas --all

echo "> Binding to ${OPENFAAS_URL}..."
kc port-forward svc/gateway -n openfaas "${OPENFAAS_PORT}":8080
