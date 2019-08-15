#!/bin/bash

# check_gofmt.sh
# Fail if a .go file hasn't been formatted with gofmt
set -euo pipefail

GO_FILES=$(find . -iname '*.go' -type f)   # All the .go files
test -z $(gofmt -s -d "$GO_FILES" | tee /dev/stderr)
