#!/bin/bash

# check_govet.sh
# Run the linter on everything except generated code
set -euo pipefail

go vet ./...
