.PHONY: build run clean test generate

# Переменные
APP_NAME = ff-split
BUILD_DIR = bin
MAIN_FILE = cmd/app/main.go

# Сборка приложения
build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)

# Запуск приложения
run: build
	./$(BUILD_DIR)/$(APP_NAME)

# Запуск в режиме разработки с hot-reload (требуется установленный air)
dev:
	air

# Очистка
clean:
	rm -rf $(BUILD_DIR)

# Запуск тестов
test:
	gotestsum --format testname ./...

# Запуск линтера
lint:
	golangci-lint run

# Старт в Docker
docker-up:
	docker-compose up -d

# Остановка Docker
docker-down:
	docker-compose down

# Миграции
migrate-up:
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/ff_split?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/ff_split?sslmode=disable" down

# Генерация кода из OpenAPI спецификации
generate:
	oapi-codegen -package api -generate types -o pkg/api/types.gen.go pkg/api/openapi.yaml
	oapi-codegen -package api -generate client -o pkg/api/client.gen.go pkg/api/openapi.yaml
	oapi-codegen -package api -generate gin-server -o pkg/api/server.gen.go pkg/api/openapi.yaml
	oapi-codegen -package api -generate spec -o pkg/api/spec.gen.go pkg/api/openapi.yaml 