#!/bin/bash

# install_dev_deps.sh
# Ensure that the tools needed to build locally are present
set -xeuo pipefail

go get github.com/golang/lint/golint
go get github.com/golang/dep/cmd/dep

GO_VERSION="go1.11beta2"
GO_SOURCE=$(go env GOPATH)/src/golang.org/x/go
mkdir -p $(dirname $GO_SOURCE)
git clone https://go.googlesource.com/go $GO_SOURCE
pushd $GO_SOURCE
git checkout $GO_VERSION
cd src && ./make.bash
popd

./scripts/get_buffalo.sh
