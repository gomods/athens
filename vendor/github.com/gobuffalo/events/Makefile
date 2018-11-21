TAGS ?= "sqlite"
GO_BIN ?= go

install:
	packr
	$(GO_BIN) install -v .
ifeq ($(GO111MODULE),on)
	$(GO_BIN) mod tidy
endif

deps:
	$(GO_BIN) get github.com/gobuffalo/release
	$(GO_BIN) get github.com/gobuffalo/packr/packr
	$(GO_BIN) get -tags ${TAGS} -t ./...
ifeq ($(GO111MODULE),on)
	$(GO_BIN) mod tidy
endif

build:
	packr
	$(GO_BIN) build -v .
ifeq ($(GO111MODULE),on)
	$(GO_BIN) mod tidy
endif

test:
	packr
	$(GO_BIN) test -tags ${TAGS} ./...
ifeq ($(GO111MODULE),on)
	$(GO_BIN) mod tidy
endif

ci-test:
	$(GO_BIN) test -tags ${TAGS} -race ./...
ifeq ($(GO111MODULE),on)
	$(GO_BIN) mod tidy
endif

lint:
	gometalinter --vendor ./... --deadline=1m --skip=internal
ifeq ($(GO111MODULE),on)
	$(GO_BIN) mod tidy
endif

update:
	$(GO_BIN) get -u -tags ${TAGS}
ifeq ($(GO111MODULE),on)
	$(GO_BIN) mod tidy
endif
	packr
	make test
	make install
ifeq ($(GO111MODULE),on)
	$(GO_BIN) mod tidy
endif

release-test:
	$(GO_BIN) test -tags ${TAGS} -race ./...
ifeq ($(GO111MODULE),on)
	$(GO_BIN) mod tidy
endif

release:
	release -y -f version.go
ifeq ($(GO111MODULE),on)
	$(GO_BIN) mod tidy
endif
