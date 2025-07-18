.PHONY: help build test run clean deps lint

# Default target
help:
	@echo "Available targets:"
	@echo "  deps        - Install dependencies"
	@echo "  build       - Build the examples"
	@echo "  test        - Run tests"
	@echo "  lint        - Run linter"
	@echo "  run         - Run simple example"
	@echo "  run-toml    - Run TOML example"
	@echo "  clean       - Clean build artifacts"

# Install dependencies
deps:
	go mod tidy
	go mod download

# Build examples
build:
	go build -o bin/simple cmd/simple/main.go
	go build -o bin/toml-example cmd/toml-example/main.go

# Run tests
test:
	go test -v ./internal/consumer/...

# Run linter
lint:
	golangci-lint run

# Run simple example
run:
	go run cmd/simple/main.go

# Run TOML example
run-toml:
	go run cmd/toml-example/main.go

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Install dependencies and build
all: deps build

# Development setup
dev-setup: deps lint test 