#!/bin/bash

# check_gofmt.sh
# Fail if a .go file hasn't been formatted with gofmt
set -euo pipefail

GO_FILES=$(find . -iname '*.go' -type f -not -path "./vendor/*")   # All the .go files
for f in $GO_FILES; do
  test -z "$(gofmt -s -w "$f" | tee /dev/stderr)"
done
