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
  rm tsoda
  find ./sql_scripts/sqlite -name *.sqlite* -delete
}
# defer cleanup, so it will be executed even after premature exit
trap cleanup EXIT

docker-compose up -d
sleep 4 # Ensure mysql is online

go build -v -tags sqlite -o tsoda ./soda

export GO111MODULE=on

function test {
  echo "!!! Testing $1"
  export SODA_DIALECT=$1
  echo ./tsoda -v
  ./tsoda drop -e $SODA_DIALECT -c ./database.yml
  ./tsoda create -e $SODA_DIALECT -c ./database.yml
  ./tsoda migrate -e $SODA_DIALECT -c ./database.yml
  go test -race -tags sqlite $verbose $(go list ./... | grep -v /vendor/)
}

test "postgres"
test "cockroach"
test "mysql"
test "sqlite"
