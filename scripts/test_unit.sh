#!/bin/bash

# test_unit.sh
# Run the unit tests with the race detector and code coverage enabled
set -xeuo pipefail

source cmd/proxy/.env
go test -race -coverprofile cover.out -covermode atomic ./...
