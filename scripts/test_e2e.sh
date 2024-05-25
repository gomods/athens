#!/bin/bash

# test_e2e.sh

# Run the e2e tests with the race detector enabled
set -xeuo pipefail
cd e2etests && go test --tags e2etests -race
