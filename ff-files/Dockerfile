# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Установка зависимостей для сборки
RUN apk add --no-cache git

# Копирование и загрузка зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/ff-files ./cmd/app

# Final stage
FROM alpine:latest

WORKDIR /app

# Копирование бинарного файла из этапа сборки
COPY --from=builder /app/ff-files .
COPY config.yaml .

# Создание директории для хранения файлов
RUN mkdir -p /app/storage && \
    adduser -D appuser && \
    chown -R appuser:appuser /app

USER appuser

EXPOSE 8082

VOLUME ["/app/storage"]

CMD ["./ff-files"] 