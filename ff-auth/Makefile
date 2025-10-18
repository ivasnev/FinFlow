# FinFlow Auth Service Makefile

# Default target
help:
	@echo "Available targets:"
	@echo "  mock-gen    - Generate mocks for services and repositories"
	@echo "  test        - Run tests"
	@echo "  build       - Build the application"
	@echo "  run         - Run the application"
	@echo "  clean       - Clean generated files"

# Generate mocks
mock-gen:
	@echo "Generating service mocks..."
	mockgen -source=internal/service/auth.go -destination=internal/service/mock/auth_mock.go -package=mock
	mockgen -source=internal/service/device.go -destination=internal/service/mock/device_mock.go -package=mock
	mockgen -source=internal/service/session.go -destination=internal/service/mock/session_mock.go -package=mock
	mockgen -source=internal/service/user.go -destination=internal/service/mock/user_mock.go -package=mock
	mockgen -source=internal/service/login_history.go -destination=internal/service/mock/login_history_mock.go -package=mock
	mockgen -source=internal/service/token.go -destination=internal/service/mock/token_mock.go -package=mock
	@echo "Generating repository mocks..."
	mockgen -source=internal/repository/device.go -destination=internal/repository/mock/device_mock.go -package=mock
	mockgen -source=internal/repository/user.go -destination=internal/repository/mock/user_mock.go -package=mock
	mockgen -source=internal/repository/session.go -destination=internal/repository/mock/session_mock.go -package=mock
	mockgen -source=internal/repository/role.go -destination=internal/repository/mock/role_mock.go -package=mock
	mockgen -source=internal/repository/login_history.go -destination=internal/repository/mock/login_history_mock.go -package=mock
	mockgen -source=internal/repository/key_pair.go -destination=internal/repository/mock/key_pair_mock.go -package=mock
	@echo "Mocks generated successfully!"

# Run tests
test:
	go test ./...

# Build the application
build:
	go build -o bin/ff-auth cmd/app/main.go

# Run the application
run:
	go run cmd/app/main.go

# Clean generated files
clean:
	rm -rf bin/
	rm -rf internal/service/mock/*.go
	rm -rf internal/repository/mock/*.go

generate:
	oapi-codegen  -package api -generate types -o pkg/api/types.gen.go pkg/api/openapi.yaml
	oapi-codegen  -package api -generate client -o pkg/api/client.gen.go pkg/api/openapi.yaml
	oapi-codegen  -package api -generate gin-server -o pkg/api/server.gen.go pkg/api/openapi.yaml
	oapi-codegen  -package api -generate spec -o pkg/api/spec.gen.go pkg/api/openapi.yaml
