#!/bin/bash

# test_unit.sh

if [ -z ${ATHENS_MONGO_CONNECTION_STRING} ]; then
    export ATHENS_MONGO_CONNECTION_STRING="mongodb://127.0.0.1:27017"
fi

if [ -z ${GO_ENV} ]; then
    export GO_ENV="test"
fi

# Run the unit tests with the race detector and code coverage enabled
set -xeuo pipefail
go test -race -coverprofile cover.out -covermode atomic ./...
