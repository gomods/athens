#!/bin/bash
#
# build.sh runs a few commands to generate build time details
# which are stored in variables in ./pkg/build

# Use the travis variables when available because travis clones different than what is on a local dev machine
# VERSION = the tag if present, otherwise the short commit hash
# BRANCH = the current branch, empty if not on a branch
if [[ "${TRAVIS-}" == "true" ]]; then
    VERSION=${TRAVIS_TAG:-${TRAVIS_COMMIT::7}}
else
    TAG=$(git describe --tags --exact-match 2> /dev/null || true)
    COMMIT=$(git rev-parse --short=7 HEAD)
    VERSION=${VERSION:-${TAG:-${COMMIT}}}
fi

DATE=$(date -u +%Y-%m-%d-%H:%M:%S-%Z)

importPath="github.com/gomods/athens/pkg/build"

export GO111MODULE=on
export CGO_ENABLED=0 

exec go build -mod=vendor -ldflags "-X $importPath.version=$VERSION -X $importPath.buildDate=$DATE" -o /bin/athens-proxy ./cmd/proxy