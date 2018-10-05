#!/bin/bash

# test_unit.sh

if [ -z ${GO_ENV} ]; then
    export GO_ENV="test"
fi

export ATHENS_MINIO_ENDPOINT="127.0.0.1:9001"

# Run the unit tests with the race detector and code coverage enabled
set -xeuo pipefail
go test -mod=vendor -race -coverprofile cover.out -covermode atomic ./...
