#!/bin/bash

# check_gofmt.sh
# Fail if a .go file hasn't been formatted with gofmt
set -euo pipefail

GO_FILES=$(find . -iname '*.go' -type f | grep -v /vendor/)   # All the .go files, excluding vendor/
gofmt_output=$(gofmt -s -l $GO_FILES | tee /dev/stderr)
echo ${gofmt_output}
test -z ${gofmt_output}
