#!/usr/bin/env bash
.PHONY: help
help: ## Display this help screen
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*?## "}; /^[a-zA-Z_-]+:.*?## / {printf "  \033[32m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# ==============================================================================
# Quality Tasks
# ==============================================================================

fmt: ## Format code
	go fmt ./...

lint: ## Run linting and vet
	go vet ./...

test: ## Run tests
	go test -v ./...

test-cover: ## Run tests with coverage
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

build: ## Build the binary
	go build -o taskhash taskhash.go

clean: ## Clean build artifacts
	rm -f taskhash coverage.out coverage.html

# ==============================================================================
# Development Tasks
# ==============================================================================

install: build ## Install taskhash globally (requires sudo)
	sudo mv taskhash /usr/local/bin/taskhash

.PHONY: help fmt lint test test-cover build clean install
