#!/usr/bin/env bash

set -xeuo pipefail

# Use the travis variables when available because travis clones different than what is on a local dev machine
# VERSION = the tag if present, otherwise the short commit hash
# BRANCH = the current branch, empty if not on a branch
if [[ "${TRAVIS-}" == "true" ]]; then
    VERSION=${TRAVIS_TAG}
else
    TAG=$(git describe --tags --exact-match 2> /dev/null || true)
    VERSION=${VERSION:-${TAG}}
fi

sed -i "s/appVersion:[^\n]*/appVersion: ${VERSION}/" charts/athens-proxy/Chart.yaml

CHART_VERSION=$( awk -F'[ .]' '/^version/ {print $1,$2"."$3"."$4+1}' ./charts/athens-proxy/Chart.yaml )
sed -i "s/version:[^\n]*/${CHART_VERSION}/" charts/athens-proxy/Chart.yaml
