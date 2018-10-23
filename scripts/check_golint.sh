#!/bin/bash

# check_golint.sh
# Run the linter on everything except generated code
set -euo pipefail

GO111MODULE=on golint -set_exit_status $(GO111MODULE=on go list ./... | grep -v '/mocks')
