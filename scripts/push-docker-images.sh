#!/bin/bash

# Push our docker images to a registry
set -xeuo pipefail

REGISTRY=${REGISTRY:-gomods/}

# Use the travis variables when available because travis clones different than what is on a local dev machine
# VERSION = the tag if present, otherwise the short commit hash
# BRANCH = the current branch, empty if not on a branch
if [[ "${TRAVIS-}" == "true" ]]; then
    VERSION=${TRAVIS_TAG:-${TRAVIS_COMMIT::7}}
    BRANCH=${TRAVIS_BRANCH}
else
    TAG=$(git describe --tags --exact-match 2> /dev/null || true)
    COMMIT=$(git rev-parse --short=7 HEAD)
    VERSION=${VERSION:-${TAG:-${COMMIT}}}
    BRANCH=${BRANCH:-$(git symbolic-ref -q --short HEAD || echo "")}
fi

# MUTABLE_TAG is the docker image tag that we will reuse between pushes, it is not a stable tag like a commit hash or tag.
if [[ "${MUTABLE_TAG:-}" == "" ]]; then
    # tagged builds
    if [[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+ ]]; then
        MUTABLE_TAG="latest"
    # master build
    elif [[ "$BRANCH" == "master" ]]; then
        MUTABLE_TAG="canary"
    # branch build
    else
        MUTABLE_TAG=${BRANCH}
    fi
fi

echo "Building Athens image for registry ${REGISTRY}"
echo "Version ${VERSION}"
echo "Mutable tag ${MUTABLE_TAG}"

REPO_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." >/dev/null && pwd )/"

echo "docker build -t ${REGISTRY}proxy:${VERSION} -f ${REPO_DIR}cmd/proxy/Dockerfile ${REPO_DIR}"
docker build -t ${REGISTRY}proxy:${VERSION} -f ${REPO_DIR}cmd/proxy/Dockerfile ${REPO_DIR}

# Apply the mutable tag to the immutable version
echo "docker tag ${REGISTRY}proxy:${VERSION} ${REGISTRY}proxy:${MUTABLE_TAG}"
docker tag ${REGISTRY}proxy:${VERSION} ${REGISTRY}proxy:${MUTABLE_TAG}

echo "docker push ${REGISTRY}proxy:${VERSION}"
docker push ${REGISTRY}proxy:${VERSION}

echo "docker push ${REGISTRY}proxy:${MUTABLE_TAG}"
docker push ${REGISTRY}proxy:${MUTABLE_TAG}
