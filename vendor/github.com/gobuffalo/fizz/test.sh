#!/bin/bash
set -e
clear

verbose=""

echo $@

if [[ "$@" == "-v" ]]
then
  verbose="-v"
fi

function cleanup {
    echo "Cleanup resources..."
    docker-compose down
    find ./sql_scripts/sqlite -name *.sqlite* -delete
}
# defer cleanup, so it will be executed even after premature exit
trap cleanup EXIT

docker-compose up -d
sleep 10 # Ensure mysql is online

go get -v -tags sqlite github.com/gobuffalo/pop/...
# go build -v -tags sqlite -o tsoda ./soda

function test {
  echo "!!! Testing $1"
  export SODA_DIALECT=$1
  soda drop -e $SODA_DIALECT
  soda create -e $SODA_DIALECT
  soda migrate -e $SODA_DIALECT
  go test -tags sqlite $verbose $(go list ./... | grep -v /vendor/)
}

test "postgres"
test "cockroach"
test "mysql"
test "sqlite"
