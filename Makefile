.PHONY: build install test test-coverage lint fmt generate clean completions docs manpages release-local audit check help

# Build variables
BINARY_NAME := shelly
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w \
	-X github.com/tj-smith47/shelly-cli/internal/version.Version=$(VERSION) \
	-X github.com/tj-smith47/shelly-cli/internal/version.Commit=$(COMMIT) \
	-X github.com/tj-smith47/shelly-cli/internal/version.Date=$(DATE) \
	-X github.com/tj-smith47/shelly-cli/internal/version.BuiltBy=make

# Default target
all: build

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -ldflags "$(LDFLAGS)" -trimpath -o bin/$(BINARY_NAME) ./cmd/shelly

## install: Install binary to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	@go install -ldflags "$(LDFLAGS)" -trimpath ./cmd/shelly

## test: Run tests
test:
	@echo "Running tests..."
	@go test -v -race ./...

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## lint: Run golangci-lint (supports FILE= for single file)
lint:
	@echo "Running linter..."
ifdef FILE
	@golangci-lint run --timeout 5m $(FILE)
else
	@golangci-lint run --timeout 5m
endif

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w -local github.com/tj-smith47/shelly-cli .

## generate: Run go generate
generate:
	@echo "Running go generate..."
	@go generate ./...

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf dist/
	@rm -f coverage.out coverage.html

## completions: Generate shell completions
completions: build
	@echo "Generating shell completions..."
	@mkdir -p completions
	@./bin/$(BINARY_NAME) completion bash > completions/shelly.bash
	@./bin/$(BINARY_NAME) completion zsh > completions/shelly.zsh
	@./bin/$(BINARY_NAME) completion fish > completions/shelly.fish
	@./bin/$(BINARY_NAME) completion powershell > completions/shelly.ps1
	@echo "Completions generated in completions/"

## docs: Generate command documentation
docs:
	@echo "Generating documentation..."
	@go run ./cmd/docgen ./docs/commands
	@echo "Documentation generated in docs/commands/"

## manpages: Generate man pages
manpages:
	@echo "Generating man pages..."
	@go run ./cmd/mangen ./docs/man
	@echo "Man pages generated in docs/man/"

## release-local: Run goreleaser locally (snapshot)
release-local:
	@echo "Running local release..."
	@goreleaser release --snapshot --clean

## deps: Download and tidy dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

## audit: Run compliance audit (RULES.md enforcement)
audit:
	@echo "Running compliance audit..."
	@./scripts/audit-compliance.sh

## check: Run all checks (lint, then compliance audit)
check:
ifdef FILE
	@$(MAKE) lint FILE=$(FILE)
else
	@$(MAKE) lint
endif
	@$(MAKE) audit
	@$(MAKE) build

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/ /'
