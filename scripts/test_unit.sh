#!/bin/bash

# test_all.sh
# Run all the tests with the race detector and code coverage enabled
set -xeuo pipefail

go test -race -coverprofile cover.out -covermode atomic ./...
