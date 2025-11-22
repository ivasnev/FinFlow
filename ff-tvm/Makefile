# FinFlow TVM Service Makefile

# Default target
help:
	@echo "Available targets:"
	@echo "  test        - Run tests with gotestsum"
	@echo "  build       - Build the application"
	@echo "  run         - Run the application"
	@echo "  clean       - Clean generated files"

# Run tests
test:
	gotestsum --format testname ./...

# Build the application
build:
	go build -o bin/ff-tvm cmd/server/main.go

# Run the application
run:
	go run cmd/server/main.go

# Clean generated files
clean:
	rm -rf bin/

