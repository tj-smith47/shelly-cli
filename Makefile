.PHONY: audit build build-test check clean completions deps docs fmt generate help install lint manpages push release-local test test-coverage

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

# Optional telemetry endpoint (set via TELEMETRY_ENDPOINT env var for production builds)
ifdef TELEMETRY_ENDPOINT
LDFLAGS += -X github.com/tj-smith47/shelly-cli/internal/telemetry.Endpoint=$(TELEMETRY_ENDPOINT)
endif

# Default target
all: build

## audit: Run compliance audit (RULES.md enforcement)
audit:
	@echo "Running compliance audit..."
	@./scripts/audit-compliance.sh

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -ldflags "$(LDFLAGS)" -trimpath -o bin/$(BINARY_NAME) ./cmd/shelly


build-test:
	@go build -o /tmp/shelly-test ./cmd/shelly

## check: Run all checks (format, lint, then compliance audit)
check:
	@$(MAKE) fmt
ifdef FILE
	@$(MAKE) lint FILE=$(FILE)
else
	@$(MAKE) lint
endif
	@$(MAKE) audit
	@$(MAKE) build

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

## deps: Download and tidy dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

## docs: Generate command documentation, man pages, and update Hugo site
docs:
	@echo "Generating command documentation..."
	@go run ./cmd/docgen ./docs/commands
	@echo "Documentation generated in docs/commands/"
	@echo "Generating man pages..."
	@go run ./cmd/mangen ./docs/man
	@echo "Man pages generated in docs/man/"
	@echo "Migrating docs to Hugo site..."
	@./scripts/migrate-docs.sh
	@./scripts/migrate-commands.sh
	@./scripts/migrate-examples.sh
	@echo "Hugo site updated in docs/site/"

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w -local github.com/tj-smith47/shelly-cli .

## generate: Run go generate
generate:
	@echo "Running go generate..."
	@go generate ./...

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/ /'

## install: Install binary to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	@go install -ldflags "$(LDFLAGS)" -trimpath ./cmd/shelly

## lint: Run golangci-lint (supports FILE= for single file)
lint:
	@echo "Running linter..."
ifdef FILE
	@golangci-lint run --timeout 5m $(FILE)
else
	@golangci-lint run --timeout 5m
endif

## manpages: Generate man pages
manpages:
	@echo "Generating man pages..."
	@go run ./cmd/mangen ./docs/man
	@echo "Man pages generated in docs/man/"

## push: Validate commit message and docs, then push
push:
	@echo "Validating commit message..."
	@COMMIT_MSG=$$(git log -1 --format=%s); \
	if ! echo "$$COMMIT_MSG" | grep -qE '#(none|patch|minor|major)'; then \
		echo "Error: Most recent commit must contain #none, #patch, #minor, or #major"; \
		echo "Commit message: $$COMMIT_MSG"; \
		exit 1; \
	fi
	@echo "Validating docs are up to date..."
	@$(MAKE) docs
	@$(MAKE) manpages
	@if ! git diff --quiet docs/; then \
		echo "Error: docs/ is out of date. Run 'make docs' and 'make manpages', then commit."; \
		git diff --stat docs/; \
		exit 1; \
	fi
	@echo "All checks passed. Pushing..."
	@git push

## release-local: Run goreleaser locally (snapshot)
release-local:
	@echo "Running local release..."
	@goreleaser release --snapshot --clean

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

## cov: Quick coverage check (shows total %)
cov:
	@go test -coverprofile=coverage.out ./... 2>&1 | tail -5
	@go tool cover -func=coverage.out | tail -1
	@rm -f coverage.out
