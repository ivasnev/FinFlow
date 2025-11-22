# FinFlow TVM Service Makefile

# Default target
help:
	@echo "Available targets:"
	@echo "  test        - Run tests with gotestsum"
	@echo "  test-cov    - Run tests with coverage"
	@echo "  build       - Build the application"
	@echo "  run         - Run the application"
	@echo "  clean       - Clean generated files"

# Run tests
test:
	gotestsum --format testname ./...

# Run tests with coverage
test-cov:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@echo ""
	@echo "Coverage report saved to coverage.out"
	@echo "To view HTML report: go tool cover -html=coverage.out"

# Build the application
build:
	go build -o bin/ff-tvm cmd/server/main.go

# Run the application
run:
	go run cmd/server/main.go

# Clean generated files
clean:
	rm -rf bin/

