#!/usr/bin/env bash

set -e

# FUNCTIONS
test_image(){
  instance=$(docker run --rm -d -p 8080:8080 "$1")
  expect="it works"
  output=$(curl -sL http://localhost:8080)
  if [[ "$output" != "$expect" ]]; then
    docker stop "${instance}"
    echo "Expected '${expect}', got '$output'"
    exit 1
  fi
  docker stop "${instance}"
}

# START
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
pushd "${DIR}"

echo "> Test with heroku/buildpacks:18"
echo ">>> Building app..."
image_name="test-app-$(openssl rand -hex 12)"
pack build "$image_name" -B heroku/buildpacks:18 -b ../ -p ./app/ --no-pull -v
echo ">>> Testing app..."
test_image "$image_name"