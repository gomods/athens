#!/bin/bash

# test_e2e.sh
# Execute end-to-end (e2e) tests to verify that everything is working right
# from the end user perpsective
set -xeuo pipefail

REPO_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )/.."

OGOPATH=${GOPATH:-}
OGO111MODULE=${GO111MODULE:-}
OGOPROXY=${GOPROXY:-}
export GO_BINARY_PATH=${GO_BINARY_PATH:-go}
TMPDIR=$(mktemp -d)
export GOPATH=$TMPDIR
GOMOD_CACHE=$TMPDIR/pkg/mod
export PATH=${REPO_DIR}/bin:${PATH}

clearGoModCache () {
  chmod -R 0770 ${GOMOD_CACHE}
  rm -fr ${GOMOD_CACHE}
}

teardown () {
  # Cleanup after our tests
  [[ -z "${OGOPATH}" ]] && unset GOPATH || export GOPATH=$OGOPATH
  [[ -z "${OGO111MODULE}" ]] && unset GO111MODULE || export GO111MODULE=$OGO111MODULE
  [[ -z "${OGOPROXY}" ]] && unset GOPROXY || export GOPROXY=$OGOPROXY

  clearGoModCache
  pkill athens-proxy || true
  rm $REPO_DIR/cmd/proxy/athens-proxy || true
  rm -fr ${TMPDIR}
  popd 2> /dev/null || true
}
trap teardown EXIT

export GO111MODULE=on

# Start the proxy in the background and wait for it to be ready
cd $REPO_DIR/cmd/proxy
pkill athens-proxy || true # cleanup proxy if it is running
go build -mod=vendor -o athens-proxy && ./athens-proxy &
while [[ "$(curl -s -o /dev/null -w ''%{http_code}'' localhost:3000)" != "200" ]]; do sleep 5; done

# Clone our test repo
TEST_SOURCE=${TMPDIR}/happy-path
rm -fr ${TEST_SOURCE} 2> /dev/null || true
git clone https://github.com/athens-artifacts/happy-path.git ${TEST_SOURCE}
pushd ${TEST_SOURCE}

# Make sure that our test repo works without the GOPROXY first
unset GOPROXY
$GO_BINARY_PATH run .

# clear cache so that go uses the proxy
clearGoModCache

# Verify that the test works against the proxy
export GOPROXY=http://localhost:3000
$GO_BINARY_PATH run .

CATALOG_RES=$(curl localhost:3000/catalog)
CATALOG_EXPECTED='{"ModsAndVersions":[{"Module":"github.com/athens-artifacts/no-tags","Version":"v0.0.0-20180803171426-1a540c5d67ab"}],"NextPageToken":""}'

if [[ "$CATALOG_RES" != "$CATALOG_EXPECTED" ]]; then
  echo ERROR: catalog endpoint failed
  exit 1 # terminate and indicate error
fi