#!/bin/bash

set +e

clear

echo "postgres"
SODA_DIALECT=postgres go test -tags sqlite -bench=.
echo "--------------------"
echo "mysql"
SODA_DIALECT=mysql go test -tags sqlite -bench=.
echo "--------------------"
echo "sqlite"
SODA_DIALECT=sqlite go test -tags sqlite -bench=.
