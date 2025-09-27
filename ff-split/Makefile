.PHONY: build run clean test

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
	go test -v ./...

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