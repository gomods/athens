#!/bin/bash
#
# build.sh runs a few commands to generate build time details
# which are stored in variables in ./pkg/build
# VERSION is expected to be set already, this is passed as a
# build argument during the call to `docker build` in
# push-docker-images.sh

DATE=$(date -u +%Y-%m-%d-%H:%M:%S-%Z)

importPath="github.com/gomods/athens/pkg/build"

export GO111MODULE=on
export CGO_ENABLED=0 

exec go build -mod=vendor -ldflags "-X $importPath.version=$VERSION -X $importPath.buildDate=$DATE" -o /bin/athens-proxy ./cmd/proxy
