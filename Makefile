# Link Checker Makefile

# Variables
BINARY_NAME=linkchecker
VERSION?=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT=$(shell git rev-parse --short HEAD)
LDFLAGS=-ldflags="-s -w -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.commit=$(COMMIT)"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt

# Build targets
.PHONY: all build clean test coverage lint fmt vet deps help install uninstall

all: clean deps test build

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) .

build-all: ## Build for all platforms
	@echo "Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 .
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe .
	GOOS=windows GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-arm64.exe .

clean: ## Clean build artifacts
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -rf dist/
	rm -f coverage.out

test: ## Run tests
	$(GOTEST) -v -race ./...

test-coverage: ## Run tests with coverage
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

coverage: test-coverage ## Alias for test-coverage

benchmark: ## Run benchmarks
	$(GOTEST) -bench=. -benchmem ./...

lint: ## Run linter
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

fmt: ## Format code
	$(GOFMT) -s -w .

fmt-check: ## Check if code is formatted
	@test -z "$$($(GOFMT) -s -l .)" || (echo "Code is not formatted. Run 'make fmt'" && exit 1)

vet: ## Run go vet
	$(GOCMD) vet ./...

deps: ## Download dependencies
	$(GOMOD) download
	$(GOMOD) tidy

deps-update: ## Update dependencies
	$(GOMOD) get -u ./...
	$(GOMOD) tidy

install: build ## Install binary to $GOPATH/bin
	cp $(BINARY_NAME) $(GOPATH)/bin/

uninstall: ## Remove binary from $GOPATH/bin
	rm -f $(GOPATH)/bin/$(BINARY_NAME)

run: ## Run the application
	$(GOCMD) run . $(ARGS)

run-example: ## Run with example arguments
	$(GOCMD) run . --help

docker-build: ## Build Docker image
	docker build -t $(BINARY_NAME):$(VERSION) .

docker-run: ## Run Docker container
	docker run --rm $(BINARY_NAME):$(VERSION)

release-dry-run: ## Simulate a release
	@echo "This would create release $(VERSION)"
	@echo "Binaries would be:"
	@echo "  - $(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz"
	@echo "  - $(BINARY_NAME)-$(VERSION)-linux-arm64.tar.gz"
	@echo "  - $(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz"
	@echo "  - $(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz"
	@echo "  - $(BINARY_NAME)-$(VERSION)-windows-amd64.zip"
	@echo "  - $(BINARY_NAME)-$(VERSION)-windows-arm64.zip"

release-local: clean build-all ## Create local release archives
	@echo "Creating release archives..."
	@mkdir -p dist/releases
	cd dist && tar -czf releases/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64
	cd dist && tar -czf releases/$(BINARY_NAME)-$(VERSION)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64
	cd dist && tar -czf releases/$(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64
	cd dist && tar -czf releases/$(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64
	cd dist && zip releases/$(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe
	cd dist && zip releases/$(BINARY_NAME)-$(VERSION)-windows-arm64.zip $(BINARY_NAME)-windows-arm64.exe
	@echo "Release archives created in dist/releases/"

check: fmt-check vet lint test ## Run all checks

ci: deps check build ## Run CI pipeline locally

# Development helpers
dev-setup: ## Set up development environment
	@echo "Setting up development environment..."
	$(GOMOD) download
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Development environment ready!"

watch: ## Watch for changes and rebuild
	@which air > /dev/null || go install github.com/cosmtrek/air@latest
	air

# Git helpers
tag: ## Create a new tag (usage: make tag VERSION=v1.0.0)
	@if [ -z "$(VERSION)" ]; then echo "Usage: make tag VERSION=v1.0.0"; exit 1; fi
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)

# Info
info: ## Show build info
	@echo "Binary Name: $(BINARY_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Commit: $(COMMIT)"
	@echo "LDFLAGS: $(LDFLAGS)" 