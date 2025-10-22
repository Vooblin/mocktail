.PHONY: build test clean install run help

# Build variables
BINARY_NAME=mocktail
BUILD_DIR=bin
MAIN_PATH=./cmd/mocktail

help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build the mocktail binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

test: ## Run all tests
	@echo "Running tests..."
	@go test ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -cover ./...

test-verbose: ## Run tests in verbose mode
	@echo "Running tests (verbose)..."
	@go test -v ./...

clean: ## Remove build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@go clean

install: ## Install mocktail to $GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	@go install $(MAIN_PATH)

run: build ## Build and run mocktail
	@$(BUILD_DIR)/$(BINARY_NAME)

deps: ## Download and tidy dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

fmt: ## Format Go code
	@echo "Formatting code..."
	@go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

lint: fmt vet ## Run all linting tools

all: clean deps lint test build ## Run all checks and build
