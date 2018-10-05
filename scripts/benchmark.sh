#!/bin/bash

go test -mod=vendor -v -bench=. $(find . -iname '*storage*test.go' -not -path '/vendor/') -run=^$
