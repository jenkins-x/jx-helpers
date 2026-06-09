#!/bin/bash

set -e -o pipefail

if [ "$DISABLE_LINTER" == "true" ]
then
  exit 0
fi

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

if ! [ -x "$(command -v golangci-lint)" ]; then
	echo "Installing GolangCI-Lint"
	${DIR}/install_golint.sh -b $GOPATH/bin v2.12.0
fi

export GO111MODULE=on
golangci-lint run \
	--no-config \
  --default=none \
  -E revive \
  -E unused \
  -E errcheck \
	-E misspell \
	-E unconvert \
  -E gosec \
  -E govet \
  -E unparam \
  -E staticcheck \
  -E goconst \
  -E ineffassign \
  -E gocritic \
  --timeout 15m0s \
  --verbose \
  --build-tags build

