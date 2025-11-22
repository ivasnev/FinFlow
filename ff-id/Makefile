# FinFlow Identity Service Makefile

# Default target
help:
	@echo "Available targets:"
	@echo "  generate    - Generate code from OpenAPI spec"
	@echo "  mock-gen    - Generate mocks for services and repositories"
	@echo "  test        - Run tests"
	@echo "  test-cov    - Run tests with coverage"
	@echo "  build       - Build the application"
	@echo "  run         - Run the application"
	@echo "  clean       - Clean generated files"

# Generate code from OpenAPI spec
generate:
	@echo "Generating code from OpenAPI spec..."
	oapi-codegen  -package api -generate types -o pkg/api/types.gen.go pkg/api/openapi.yaml
	oapi-codegen  -package api -generate client -o pkg/api/client.gen.go pkg/api/openapi.yaml
	oapi-codegen  -package api -generate gin-server -o pkg/api/server.gen.go pkg/api/openapi.yaml
	oapi-codegen  -package api -generate spec -o pkg/api/spec.gen.go pkg/api/openapi.yaml
	@echo "Code generation completed successfully!"

# Generate mocks
mock-gen:
	@echo "Generating service mocks..."
	mockgen -source=internal/service/interfaces.go -destination=internal/service/mock/interfaces_mock.go -package=mock
	@echo "Generating repository mocks..."
	mockgen -source=internal/repository/interfaces.go -destination=internal/repository/mock/interfaces_mock.go -package=mock
	@echo "Mocks generated successfully!"

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
	go build -o bin/ff-id cmd/app/main.go

# Run the application
run:
	go run cmd/app/main.go

# Clean generated files
clean:
	rm -rf bin/
	rm -rf internal/service/mock/*.go
	rm -rf internal/repository/mock/*.go
	rm -rf pkg/api/*.gen.go
