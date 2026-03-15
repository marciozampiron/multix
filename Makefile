# Multix - Enterprise Multi-Cloud CLI

BINARY_NAME:=multix
VERSION?=1.0.0-beta
BUILD_DIR:=build
GO:=go
GOFLAGS:=-ldflags "-s -w -X multix/pkg/version.Version=$(VERSION)"

.PHONY: all build run test test-race lint fmt vet vuln tidy clean install

all: clean fmt vet tidy test build

build: ## Build the Enterprise CLI binary
	@echo "==> Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/multix/main.go
	@echo "==> Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

run: build ## Build and run the CLI (usage: make run ARGS="cloud list")
	@$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

test: ## Run unit tests
	@echo "==> Running tests..."
	@$(GO) test -v ./... -cover

test-race: ## Run tests with race detector
	@echo "==> Running tests with race detector..."
	@$(GO) test -v ./... -race -cover

fmt: ## Run go fmt
	@echo "==> Formatting code..."
	@$(GO) fmt ./...

vet: ## Run go vet
	@echo "==> Running go vet..."
	@$(GO) vet ./...

tidy: ## Clean up go.mod and go.sum
	@echo "==> Tidy go modules..."
	@$(GO) mod tidy

clean: ## Remove build artifacts
	@echo "==> Cleaning up..."
	@rm -rf $(BUILD_DIR)

install: build ## Install the binary to GOPATH/bin
	@echo "==> Installing $(BINARY_NAME) to $(GOPATH)/bin..."
	@$(GO) install $(GOFLAGS) cmd/multix/main.go

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
