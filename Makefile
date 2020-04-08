# A lot of this is a straight rip-off from @jessfraz, because it's that awesome.

# Set an output prefix, which is the local directory if not specified
PREFIX?=$(shell pwd)

.PHONY: all
all: help

.PHONY: build
build: ## Build compote.
	@echo "+ $@"
	@go build -o compote -ldflags="-s -w" main.go

.PHONY: container
container: ## Build the compote container.
	@echo "+ $@"
	@./build.sh

.PHONY: fmt
fmt: ## Verifies all files have been `gofmt`ed.
	@echo "+ $@"
	@gofmt -s -l . | grep -v '.pb.go:' | grep -v vendor | tee /dev/stderr

.PHONY: lint
lint: ## Verifies `golint` passes.
	@echo "+ $@"
	@golint ./... | grep -v '.pb.go:' | grep -v vendor | tee /dev/stderr

.PHONY: test
test: ## Run the default tests for this project.
	@echo "+ $@"
	@go test ${GO_EXTRAS} -v $(shell go list ./... | grep -v vendor)

.PHONY: test-c
test-c: ## Run the default tests for this project and show coverage.
	@echo "+ $@"
	@GO_EXTRAS="-cover -coverprofile=c.out" make test
	@go tool cover -html=c.out

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

