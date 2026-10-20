# Этап сборки
FROM golang:1.23-alpine AS builder

# Устанавливаем git для скачивания Go модулей
RUN apk add --no-cache git

# Сборка из корня репозитория сохраняет локальные исправления общих модулей.
WORKDIR /src
COPY ff-id/ ./ff-id/
COPY ff-auth/ ./ff-auth/
COPY ff-tvm/ ./ff-tvm/
WORKDIR /src/ff-id
RUN go mod download

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/app

# Финальный этап
FROM alpine:latest

WORKDIR /app

# Копируем бинарный файл из этапа сборки
COPY --from=builder /app/server .

COPY ff-id/config/ ./config

# Открываем порт
EXPOSE 8083

# Запускаем приложение
CMD ["./server"]