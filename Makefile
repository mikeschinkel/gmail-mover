# Gmail Mover Makefile
# Provides common development tasks for building, testing, and running Gmail Mover

.PHONY: all build test test-integration clean fmt vet tidy install help run dev deps check

# Default target
all: build test

# Variables
BINARY_NAME=gmail-mover
BINARY_PATH=bin/$(BINARY_NAME)
CMD_PATH=./cmd
TEST_TIMEOUT=30s

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
BOLD=\033[1m
NC=\033[0m # No Color

## Build Commands

# Build the main binary
build:
	@echo "$(BLUE)🔨 Building $(BINARY_NAME)...$(NC)"
	@mkdir -p bin
	@go -C $(CMD_PATH) build -o ../$(BINARY_PATH)
	@echo "$(GREEN)✅ Binary built successfully: $(BINARY_PATH)$(NC)"

# Install binary to GOPATH/bin
install: build
	@echo "$(BLUE)📦 Installing $(BINARY_NAME) to GOPATH/bin...$(NC)"
	@go -C $(CMD_PATH) install
	@echo "$(GREEN)✅ Installation completed$(NC)"

# Clean build artifacts
clean:
	@echo "$(YELLOW)🧹 Cleaning build artifacts...$(NC)"
	@rm -rf bin/
	@go clean
	@go -C $(CMD_PATH) clean
	@echo "$(GREEN)✅ Clean completed$(NC)"

## Testing Commands

# Run all tests
test: test-integration
	@echo "$(GREEN)🎉 All tests completed!$(NC)"

# Run integration tests (test package)
test-integration: build
	@echo "$(BLUE)🧪 Running integration tests...$(NC)"
	@go test -v -timeout $(TEST_TIMEOUT) ./test/...
	@echo "$(GREEN)✅ Integration tests completed$(NC)"

# Run tests with coverage
test-coverage:
	@echo "$(BLUE)📊 Running tests with coverage...$(NC)"
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✅ Coverage report generated: coverage.html$(NC)"

# Run tests in watch mode (requires entr)
test-watch:
	@echo "$(BLUE)👀 Running tests in watch mode...$(NC)"
	@echo "$(YELLOW)Press Ctrl+C to stop$(NC)"
	@find . -name "*.go" | entr -r make test

## Code Quality Commands

# Format all Go code
fmt:
	@echo "$(BLUE)📝 Formatting Go code...$(NC)"
	@go fmt ./...
	@go -C $(CMD_PATH) fmt ./...
	@echo "$(GREEN)✅ Code formatted$(NC)"

# Vet code for issues
vet:
	@echo "$(BLUE)🔍 Vetting code...$(NC)"
	@go vet ./...
	@go -C $(CMD_PATH) vet ./...
	@echo "$(GREEN)✅ Code vetted$(NC)"

# Tidy dependencies
tidy:
	@echo "$(BLUE)📦 Tidying dependencies...$(NC)"
	@go mod tidy
	@go -C $(CMD_PATH) mod tidy
	@go work sync
	@echo "$(GREEN)✅ Dependencies tidied$(NC)"

# Download dependencies
deps:
	@echo "$(BLUE)📦 Downloading dependencies...$(NC)"
	@go mod download
	@go -C $(CMD_PATH) mod download
	@echo "$(GREEN)✅ Dependencies downloaded$(NC)"

# Run all code quality checks
check: fmt vet tidy
	@echo "$(GREEN)✅ All code quality checks completed$(NC)"

## Development Commands

# Run in development mode
dev: build
	@echo "$(BLUE)🚀 Starting Gmail Mover in development mode...$(NC)"
	@echo "$(YELLOW)Use --help to see available options$(NC)"
	@./$(BINARY_PATH) --help

# Run with source and destination emails
run: build
	@echo "$(BLUE)🚀 Running Gmail Mover...$(NC)"
	@if [ -z "$(SOURCE)" ] || [ -z "$(DEST)" ]; then \
		echo "$(RED)❌ Usage: make run SOURCE=source@gmail.com DEST=dest@gmail.com [OPTIONS]$(NC)"; \
		echo "$(YELLOW)Available OPTIONS: MAX=100 DRY_RUN=true QUERY='from:sender'$(NC)"; \
		exit 1; \
	fi
	@./$(BINARY_PATH) --source=$(SOURCE) --dest=$(DEST) \
		$(if $(MAX),--max=$(MAX)) \
		$(if $(DRY_RUN),--dry-run) \
		$(if $(QUERY),--query="$(QUERY)")

# Run with job file
run-job: build
	@echo "$(BLUE)🚀 Running Gmail Mover with job file...$(NC)"
	@if [ -z "$(JOB)" ]; then \
		echo "$(RED)❌ Usage: make run-job JOB=path/to/job.json$(NC)"; \
		exit 1; \
	fi
	@./$(BINARY_PATH) --job=$(JOB)

# List labels for an email account
list-labels: build
	@echo "$(BLUE)📋 Listing labels for Gmail account...$(NC)"
	@if [ -z "$(EMAIL)" ]; then \
		echo "$(RED)❌ Usage: make list-labels EMAIL=user@gmail.com$(NC)"; \
		exit 1; \
	fi
	@./$(BINARY_PATH) --list-labels=$(EMAIL)

# Run in dry-run mode for testing
dry-run: build
	@echo "$(BLUE)🧪 Running Gmail Mover in dry-run mode...$(NC)"
	@if [ -z "$(SOURCE)" ] || [ -z "$(DEST)" ]; then \
		echo "$(RED)❌ Usage: make dry-run SOURCE=source@gmail.com DEST=dest@gmail.com$(NC)"; \
		exit 1; \
	fi
	@./$(BINARY_PATH) --source=$(SOURCE) --dest=$(DEST) --dry-run --max=5

## Help

# Show help
help:
	@echo "$(BOLD)Gmail Mover Development Commands$(NC)"
	@echo ""
	@echo "$(BOLD)Build Commands:$(NC)"
	@echo "  $(BLUE)build$(NC)         Build the gmail-mover binary"
	@echo "  $(BLUE)install$(NC)       Install binary to GOPATH/bin"
	@echo "  $(BLUE)clean$(NC)         Clean build artifacts"
	@echo ""
	@echo "$(BOLD)Testing Commands:$(NC)"
	@echo "  $(BLUE)test$(NC)          Run all tests (integration)"
	@echo "  $(BLUE)test-integration$(NC) Run integration tests only"
	@echo "  $(BLUE)test-coverage$(NC) Run tests with coverage report"
	@echo "  $(BLUE)test-watch$(NC)    Run tests in watch mode (requires entr)"
	@echo ""
	@echo "$(BOLD)Code Quality:$(NC)"
	@echo "  $(BLUE)fmt$(NC)           Format all Go code"
	@echo "  $(BLUE)vet$(NC)           Vet code for issues"
	@echo "  $(BLUE)tidy$(NC)          Tidy dependencies"
	@echo "  $(BLUE)deps$(NC)          Download dependencies"
	@echo "  $(BLUE)check$(NC)         Run all code quality checks"
	@echo ""
	@echo "$(BOLD)Development:$(NC)"
	@echo "  $(BLUE)dev$(NC)           Show help and development info"
	@echo "  $(BLUE)run$(NC)           Run with emails: make run SOURCE=s@gmail.com DEST=d@gmail.com"
	@echo "  $(BLUE)run-job$(NC)       Run with job file: make run-job JOB=job.json"
	@echo "  $(BLUE)list-labels$(NC)   List labels: make list-labels EMAIL=user@gmail.com"
	@echo "  $(BLUE)dry-run$(NC)       Test run: make dry-run SOURCE=s@gmail.com DEST=d@gmail.com"
	@echo ""
	@echo "$(BOLD)Examples:$(NC)"
	@echo "  make build"
	@echo "  make test"
	@echo "  make run SOURCE=main@gmail.com DEST=archive@gmail.com MAX=50"
	@echo "  make run-job JOB=examples/daily-archive.json"
	@echo "  make list-labels EMAIL=main@gmail.com"
	@echo "  make dry-run SOURCE=main@gmail.com DEST=archive@gmail.com"