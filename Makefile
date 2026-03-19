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
	go test -race ./...

test-cover: ## Run tests with coverage
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

build: ## Build the binary
	go build -o taskhash taskhash.go

clean: ## Clean build artifacts
	rm -f taskhash coverage.out coverage.html *.prof

# ==============================================================================
# Development Tasks
# ==============================================================================

install: build ## Install taskhash globally (requires sudo)
	sudo mv taskhash /usr/local/bin/taskhash

prof: ## Run CPU and memory profiling
	@echo "Running benchmarks and generating profiles..."
	go test -bench=. -benchtime=500ms -cpuprofile=cpu.prof -memprofile=mem.prof ./...
	@echo ""
	@echo "Profiles generated: cpu.prof, mem.prof"
	@echo "Analyze with:"
	@echo "  go tool pprof cpu.prof    # CPU analysis"
	@echo "  go tool pprof mem.prof    # Memory analysis"
	@echo ""
	@echo "Inside pprof:"
	@echo "  top 10        # Top consumers"
	@echo "  list <func>   # Source for function"
	@echo "  web           # Open browser (if available)"

.PHONY: help fmt lint test test-cover build clean install prof
