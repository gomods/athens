VERSION = "unset"
DATE=$(shell date -u +%Y-%m-%d-%H:%M:%S-%Z)

GOLANGCI_LINT_VERSION=v1.51.2

ifndef GOLANG_VERSION
override GOLANG_VERSION = 1.20
endif

.PHONY: build
build: ## build the athens proxy
	go build -ldflags="-w -s" -o ./cmd/proxy/proxy ./cmd/proxy

.PHONY: build-ver
build-ver: ## build the athens proxy with version number
	GO111MODULE=on CGO_ENABLED=0 GOPROXY="https://proxy.golang.org" go build -ldflags "-s -w -X github.com/gomods/athens/pkg/build.version=$(VERSION) -X github.com/gomods/athens/pkg/build.buildDate=$(DATE)" -o athens ./cmd/proxy

athens:
	$(MAKE) build-ver

# The build-image step creates a docker image that has all the tools required
# to perform some CI build steps, instead of relying on them being installed locally
.PHONY: build-image
build-image:
	docker build -t athens-build ./scripts/build-image

.PHONY: run
run: ## run the athens proxy with dev configs
	cd ./cmd/proxy && go run . -config_file ../../config.dev.toml

.PHONY: run-docker
run-docker:
	docker-compose -p athensdockerdev build --build-arg GOLANG_VERSION=${GOLANG_VERSION} dev
	docker-compose -p athensdockerdev up -d dev

.PHONY: run-docker-teardown
run-docker-teardown:
	docker-compose -p athensdockerdev down

.PHONY: docs
docs: ## build the docs docker image
	docker build -t gomods/hugo -f docs/Dockerfile .

.PHONY: docs-docker
docs-docker: ## run the docs docker image
	docker run -it --rm --name hugo-server -p 1313:1313 -v ${PWD}/docs:/src:cached gomods/hugo

.PHONY: setup-dev-env
setup-dev-env:
	$(MAKE) dev

.PHONY: lint
lint:
	@golangci-lint run ./...

.PHONY: lint-docker
lint-docker:
	@docker run -t --rm -v $(CURDIR):/app -w /app golangci/golangci-lint:$(GOLANGCI_LINT_VERSION) golangci-lint run ./...

.PHONY: verify
verify: ## verify athens codebase
	./scripts/check_deps.sh
	./scripts/check_conflicts.sh

.PHONY: test-unit
test-unit: ## run unit tests with race detector and code coverage enabled
	./scripts/test_unit.sh

.PHONY: test-unit-docker
test-unit-docker: ## run unit tests with docker
	docker-compose -p athensunit build --build-arg GOLANG_VERSION=${GOLANG_VERSION} testunit	
	docker-compose -p athensunit up --exit-code-from=testunit testunit
	docker-compose -p athensunit down

.PHONY: test-e2e
test-e2e:
	./scripts/test_e2e.sh

.PHONY: test-e2e-docker
test-e2e-docker:
	docker-compose -p athense2e build --build-arg GOLANG_VERSION=${GOLANG_VERSION} teste2e
	docker-compose -p athense2e up --exit-code-from=teste2e teste2e
	docker-compose -p athense2e down

.PHONY: docker
docker: proxy-docker

.PHONY: proxy-docker
proxy-docker:
	docker build -t gomods/athens -f cmd/proxy/Dockerfile --build-arg GOLANG_VERSION=${GOLANG_VERSION} .

.PHONY: docker-push
docker-push:
	./scripts/push-docker-images.sh

bench:
	./scripts/benchmark.sh

.PHONY: alldeps
alldeps:
	docker-compose -p athensdev up -d mongo
	docker-compose -p athensdev up -d minio
	docker-compose -p athensdev up -d jaeger
	echo "sleeping for a bit to wait for the DB to come up"
	sleep 5

.PHONY: dev
dev:
	docker-compose -p athensdev up -d mongo
	docker-compose -p athensdev up -d minio

.PHONY: down
down:
	docker-compose -p athensdev down -v

.PHONY: dev-teardown
dev-teardown:
	docker-compose -p athensdev down -v

.PHONY: clean
clean: ## delete all locally-built artefacts (not including docker images)
	rm -f athens cmd/proxy/proxy

.PHONY: help
help: ## display help page
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: deploy-gae
deploy-gae:
	cd scripts/gae && gcloud app deploy
