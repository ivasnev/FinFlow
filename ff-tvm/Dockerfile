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
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/ff-tvm ./cmd/app

# Final stage
FROM alpine:latest

WORKDIR /app

# Копирование бинарного файла из этапа сборки
COPY --from=builder /app/ff-tvm .
COPY config.yaml .

# Создание пользователя без прав root
RUN adduser -D appuser && \
    chown -R appuser:appuser /app

USER appuser

EXPOSE 8083

CMD ["./ff-tvm"] 