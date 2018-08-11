#!/bin/bash

# test_unit.sh

source cmd/proxy/.env
export TMPDIR=$(mktemp -d)

# Run the unit tests with the race detector and code coverage enabled
set -xeuo pipefail
go test -race -coverprofile cover.out -covermode atomic ./...
