BINARY=gandalf
BUILD_DIR=bin
GO_FILES=$(shell find . -type f -name '*.go' | grep -v vendor)

DIST_DIR=dist
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build run test tidy clean mellon lint fmt fmt-check vet check dist dist-clean

build:
	mkdir -p $(BUILD_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY) ./cmd/gandalf

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
	rm -rf $(BUILD_DIR) $(DIST_DIR)

# Remove only dist artifacts
dist-clean:
	rm -rf $(DIST_DIR)

# Cross-compile for all supported platforms (Linux, macOS, Windows) amd64/arm64
dist:
	@echo "Building $(VERSION) for all platforms..."
	mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/gandalf-linux-amd64 ./cmd/gandalf
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/gandalf-linux-arm64 ./cmd/gandalf
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/gandalf-darwin-amd64 ./cmd/gandalf
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/gandalf-darwin-arm64 ./cmd/gandalf
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/gandalf-windows-amd64.exe ./cmd/gandalf
	GOOS=windows GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/gandalf-windows-arm64.exe ./cmd/gandalf
	cd $(DIST_DIR) && sha256sum gandalf-* > checksums.txt 2>/dev/null || shasum -a 256 gandalf-* > checksums.txt
	@echo "Artifacts in $(DIST_DIR)/:"
	@ls -lh $(DIST_DIR)/
	@cat $(DIST_DIR)/checksums.txt

# Create archives for release (tar.gz for Unix, zip for Windows)
dist-archives: dist
	cd $(DIST_DIR) && tar -czf gandalf-linux-amd64.tar.gz gandalf-linux-amd64 && tar -czf gandalf-linux-arm64.tar.gz gandalf-linux-arm64
	cd $(DIST_DIR) && tar -czf gandalf-darwin-amd64.tar.gz gandalf-darwin-amd64 && tar -czf gandalf-darwin-arm64.tar.gz gandalf-darwin-arm64
	cd $(DIST_DIR) && zip -j gandalf-windows-amd64.zip gandalf-windows-amd64.exe >/dev/null && zip -j gandalf-windows-arm64.zip gandalf-windows-arm64.exe >/dev/null
	cd $(DIST_DIR) && sha256sum *.tar.gz *.zip > checksums.txt 2>/dev/null || shasum -a 256 *.tar.gz *.zip > checksums.txt
	@ls -lh $(DIST_DIR)/*.tar.gz $(DIST_DIR)/*.zip

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/gandalf

help:
	@echo "Available targets:"
	@echo "  build        - Build binary to $(BUILD_DIR)/$(BINARY)"
	@echo "  dist         - Cross-compile for Linux/macOS/Windows (amd64/arm64) to $(DIST_DIR)/"
	@echo "  dist-archives- Create tar.gz/zip archives + checksums for release"
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
	@echo "  clean        - Remove build artifacts (bin + dist)"
