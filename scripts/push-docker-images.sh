#!/bin/bash

# Push our docker images to a registry
set -xeuo pipefail

REGISTRY=${REGISTRY:-gomods/}
TAG=$(git describe --tags --exact-match 2> /dev/null || true)
COMMIT=$(git rev-parse --short=7 HEAD)
VERSION=${VERSION:-${TAG:-${COMMIT}}}
BRANCH=${BRANCH:-$(git symbolic-ref -q --short HEAD || echo "")}

# MUTABLE_TAG is the docker image tag that we will reuse between pushes, it is not an immutable tag like a commit hash or tag.
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

REPO_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." >/dev/null && pwd )/"

docker build --build-arg VERSION=${VERSION} -t ${REGISTRY}athens:${VERSION} -f ${REPO_DIR}cmd/proxy/Dockerfile ${REPO_DIR}

# Apply the mutable tag to the immutable version
docker tag ${REGISTRY}athens:${VERSION} ${REGISTRY}athens:${MUTABLE_TAG}

docker push ${REGISTRY}athens:${VERSION}
docker push ${REGISTRY}athens:${MUTABLE_TAG}
