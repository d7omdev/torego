# Variables
BINARY_NAME = torego
CMD_PATH = ./cmd/torego

# Default target
.DEFAULT_GOAL := help

# Commands
.PHONY: init build  run clean deps help

# Initialize the go.mod file and download dependencies
init:
	@echo "Initializing Go module and installing dependencies..."
	@go mod init github.com/yourusername/torego || true
	@$(MAKE) deps

# Download and install dependencies
deps:
	@echo "Installing dependencies..."
	@go get -u ./...
	@go mod tidy
	@echo "Dependencies installed."

# Build the torego binary
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o build/$(BINARY_NAME) $(CMD_PATH)
	@echo "$(BINARY_NAME) built successfully!"

# Run the torego binary
run: clean build
	@echo "Running $(BINARY_NAME)..."
	@echo ""
	@build/$(BINARY_NAME)

# Clean up the build and temporary files
clean:
	@echo "Cleaning up..."
	@rm -f $(BINARY_NAME)
	@rm -f torego.db
	@echo "Cleaned up."

# Show help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  init      Initialize the Go module and install dependencies"
	@echo "  deps      Install dependencies"
	@echo "  build     Build the torego binary"
	@echo "  run       Run the torego binary"
	@echo "  clean     Clean up the build and temporary files"
	@echo "  help      Show this help message"
	@echo ""
