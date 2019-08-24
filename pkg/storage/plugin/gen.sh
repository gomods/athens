#!/bin/env bash

set -ve
shopt -s globstar

protoc \
    -I/usr/include \
    -I/usr/local/include \
    -I/usr/bin \
    -I./pb \
    --go_out=plugins=grpc,paths=source_relative:./pb \
    ./pb/**/*.proto
