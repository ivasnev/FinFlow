# Этап сборки
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Копируем файлы с зависимостями
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server

# Финальный этап
FROM alpine:latest

WORKDIR /app

# Копируем бинарный файл из этапа сборки
COPY --from=builder /app/server .

# Копируем миграции
COPY migrations ./migrations
COPY config/ ./config

# Открываем порт
EXPOSE 8080

# Запускаем приложение
CMD ["./server"] 