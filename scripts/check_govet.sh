#!/bin/bash

# check_golint.sh
# Run the linter on everything except generated code
set -euo pipefail

go vet ./...
