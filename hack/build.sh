#!/usr/bin/env bash

set -e

TAG=${1:-dev}
VERSION=$(git describe --tags --abbrev=0)
COMMIT=$(git rev-parse --short HEAD)
BINARY=mtg."${TAG}"

CONFIG=config."${TAG}".yaml
if [ -f "${CONFIG}" ]; then
  trap 'rm -f config_gen.go' EXIT
  if ! type "config-gen" > /dev/null 2>/dev/null; then
    env GO111MODULE=off go get -u github.com/fox-one/pkg/config/config-gen
  fi
  echo "use config ${CONFIG}"
  config-gen --config "${CONFIG}" --tag "${TAG}"
fi

export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0

echo "build ${BINARY} with version ${VERSION} & commit ${COMMIT}"
go build -a \
         --tags "${TAG}" \
         --ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" \
         -o builds/${BINARY}

trap 'rm -f config_gen.go' EXIT
