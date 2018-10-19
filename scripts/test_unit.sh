#!/bin/bash

# test_unit.sh

if [ -z ${GO_ENV} ]; then
    export GO_ENV="test"
fi

export GO111MODULE=on

# Run the unit tests with the race detector and code coverage enabled
set -xeuo pipefail
go test -mod=readonly -race -coverprofile cover.out -covermode atomic ./...
