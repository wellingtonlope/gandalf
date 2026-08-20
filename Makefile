BINARY=gandalf
BUILD_DIR=bin
GO_FILES=$(shell find . -type f -name '*.go' | grep -v vendor)

.PHONY: build run test tidy clean mellon lint fmt fmt-check vet check

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/gandalf

run: build
	./$(BUILD_DIR)/$(BINARY)

mellon: build
	./$(BUILD_DIR)/$(BINARY) mellon

# Run all unit tests with race detector and coverage
test:
	go test ./... -count=1 -race -cover

# Run tests verbosely (useful for debugging)
test-verbose:
	go test ./... -count=1 -race -v

# Format all Go files
fmt:
	go fmt ./...
	gofmt -s -w .

# Check if files are formatted (fails if not formatted)
fmt-check:
	@test -z "$$(gofmt -l .)" || (echo "Files not formatted. Run 'make fmt'." && gofmt -l . && exit 1)
	@test -z "$$(go fmt ./... 2>&1 | grep -v '^$$')" || (echo "go fmt needed" && exit 1)

# Run linter: prefer golangci-lint if available, fallback to go vet
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "Running golangci-lint..."; \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found, falling back to 'go vet'..."; \
		echo "Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		go vet ./...; \
	fi

# Alias for lint (common typo)
golinter: lint

# Run go vet only
vet:
	go vet ./...

# Full pre-commit check: format, lint, test
check: fmt-check lint test
	@echo "All checks passed! Ready to commit."

# Also run build to ensure it compiles after checks
verify: check build
	@echo "Verify complete: formatted, linted, tested and built."

tidy:
	go mod tidy

clean:
	rm -rf $(BUILD_DIR)

install:
	go install ./cmd/gandalf

help:
	@echo "Available targets:"
	@echo "  build        - Build binary to $(BUILD_DIR)/$(BINARY)"
	@echo "  test         - Run tests with race and coverage"
	@echo "  test-verbose - Run tests verbosely"
	@echo "  fmt          - Format code (go fmt + gofmt)"
	@echo "  fmt-check    - Check if code is formatted"
	@echo "  lint         - Run golangci-lint (or go vet fallback)"
	@echo "  vet          - Run go vet"
	@echo "  check        - Pre-commit: fmt-check + lint + test"
	@echo "  verify       - check + build"
	@echo "  mellon       - Build and reveal the Doors of Durin"
	@echo "  tidy         - go mod tidy"
	@echo "  clean        - Remove build artifacts"
