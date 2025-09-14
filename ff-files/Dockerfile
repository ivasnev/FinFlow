FROM golang:1.23-alpine AS builder

WORKDIR /app

# Копируем файлы проекта
COPY . .

# Собираем приложение
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server

# Создаем финальный образ
FROM alpine:latest

WORKDIR /app

# Копируем бинарный файл из builder
COPY --from=builder /app/server .

# Запускаем приложение
CMD ["./server"] 