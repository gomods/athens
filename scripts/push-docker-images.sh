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
    # main branch build
    elif [[ "$BRANCH" == "main" ]]; then
        MUTABLE_TAG="canary"
    # branch build
    else
        MUTABLE_TAG=${BRANCH}
    fi
fi

REPO_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." >/dev/null && pwd )/"

# Build & Publish Multi-arch Docker Image
docker buildx build \
  --pull --push \
  --platform=linux/amd64/v1,linux/arm64/v8 \
  --tag "${REGISTRY}athens:${VERSION}" \
  --tag "${REGISTRY}athens:${MUTABLE_TAG}" \
  --build-arg=VERSION=${VERSION} \
  --file ${REPO_DIR}cmd/proxy/Dockerfile ${REPO_DIR}

