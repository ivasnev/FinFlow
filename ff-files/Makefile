# FinFlow Files Service Makefile

# Default target
help:
	@echo "Available targets:"
	@echo "  generate    - Generate API code from OpenAPI spec"
	@echo "  test        - Run tests with gotestsum"
	@echo "  build       - Build the application"
	@echo "  run         - Run the application"
	@echo "  clean       - Clean generated files"

# Generate API code from OpenAPI spec
generate:
	@echo "Generating API code from OpenAPI spec..."
	oapi-codegen -package api -generate types -o pkg/api/types.gen.go pkg/api/openapi.yaml
	oapi-codegen -package api -generate client -o pkg/api/client.gen.go pkg/api/openapi.yaml
	oapi-codegen -package api -generate gin-server -o pkg/api/server.gen.go pkg/api/openapi.yaml
	oapi-codegen -package api -generate spec -o pkg/api/spec.gen.go pkg/api/openapi.yaml
	@echo "API code generated successfully!"

# Run tests
test:
	gotestsum --format testname ./...

# Build the application
build:
	go build -o bin/ff-files cmd/server/main.go

# Run the application
run:
	go run cmd/server/main.go

# Clean generated files
clean:
	rm -rf bin/
	rm -rf pkg/api/*.gen.go
