.PHONY: build run test clean

# Build the application
build:
	go build -o bin/ai-dev-env ./cmd/ai-dev-env

# Run the application
run: build
	./bin/ai-dev-env

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Install dependencies
deps:
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	go vet ./...

# Default target
all: clean build