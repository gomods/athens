# Run the unit tests with the race detector and code coverage enabled

if (!(Test-Path env:GO_ENV)) {$env:GO_ENV = "test"}

& go test -mod=vendor -race -coverprofile cover.out -covermode atomic ./...
