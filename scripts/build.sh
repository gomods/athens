#!/bin/bash
#
# build.sh runs a few commands to generate build time details
# which are stored in variables in ./pkg/build

commit=$(git rev-list -1 HEAD)
version=$(git describe --tags)
date=$(date -u +%Y-%m-%d-%H:%M:%S-%Z)

importPath="github.com/gomods/athens/pkg/build"

export GO111MODULE=on
export CGO_ENABLED=0 

exec go build -mod=vendor -ldflags "-X $importPath.commitSHA=$commit -X $importPath.version=$version -X $importPath.buildDate=$date" -o /bin/athens-proxy ./cmd/proxy