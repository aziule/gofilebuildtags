# Variables
GOLANGCI-LINT-VERSION=v1.62.0
GOLANGCI-LINT=docker run --env=GOFLAGS=-mod=vendor --rm -v $(CURDIR):/app -w /app golangci/golangci-lint:$(GOLANGCI-LINT-VERSION) golangci-lint

# Targets
default: help

build: ## Build from source
	@go build -o filebuildtag ./cmd/filebuildtag/

cover: ## Run unit tests coverage
	@go test -race -failfast -count=1 -coverprofile=coverage.out .
	@go tool cover -html=coverage.out

lint: ## Lint Go files (Docker must be up)
	@$(GOLANGCI-LINT) run

test: ## Run unit tests
	@go test -race -failfast -count=1 .

help:
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":[^:]*?## "}; {printf "\t\033[0;34m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: build cover lint test help