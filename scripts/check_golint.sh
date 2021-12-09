#!/bin/bash

# check_golint.sh
# Run the linter on everything except generated code
set -euo pipefail

golint -set_exit_status $(GO111MODULE=auto go list ./... | grep -v '/mocks')
