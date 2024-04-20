# Run the unit tests with the race detector and code coverage enabled

if (!(Test-Path env:GO_ENV)) {$env:GO_ENV = "test"}

if (!(Test-Path env:ATHENS_MINIO_ENDPOINT)) {
    $env:ATHENS_MINIO_ENDPOINT = "http://127.0.0.1:9001"
}

if (!(Test-Path env:ATHENS_MONGO_STORAGE_URL)) {
    $env:ATHENS_MONGO_STORAGE_URL = "mongodb://127.0.0.1:27017"
}
$env:GO111MODULE="on"
& go test -mod=vendor -race -coverprofile cover.out -covermode atomic ./...
