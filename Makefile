.PHONY: build test test-race test-cover test-cover-html lint fmt vet staticcheck check clean docker-up docker-down run-example-with-jaeger

# Default go command
GO ?= go

# Directories to analyze
DIRS ?= ./...

# Build the package
build:
	$(GO) build -v $(DIRS)

# Run all tests
test:
	$(GO) test -v $(DIRS)

# Run tests with race detection
test-race:
	$(GO) test -race -v $(DIRS)

# Run tests with coverage
test-cover:
	$(GO) test -coverprofile=coverage.out $(DIRS)

# Generate HTML coverage report
test-cover-html: test-cover
	$(GO) tool cover -html=coverage.out

# Run go fmt
fmt:
	$(GO) fmt $(DIRS)

# Run go vet
vet:
	$(GO) vet $(DIRS)

# Run all linters
lint: fmt vet

# Run tests and linting
check: test lint

# Clean up
clean:
	rm -f coverage.out
	rm -rf ./dist

# Docker Compose up
docker-up:
	cd example && docker-compose up -d

# Docker Compose down
docker-down:
	cd example && docker-compose down

# Run the example with Jaeger (starts Docker Compose first)
run-example: docker-up
	cd example && $(GO) run main.go
