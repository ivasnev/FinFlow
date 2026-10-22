# Этап сборки
FROM golang:1.27.3-alpine AS builder

# Устанавливаем git для скачивания Go модулей
RUN apk add --no-cache git

# Сборка из корня сохраняет локальную проверку назначения токенов.
WORKDIR /src
COPY ff-split/ ./ff-split/
COPY ff-auth/ ./ff-auth/
COPY ff-id/ ./ff-id/
COPY ff-tvm/ ./ff-tvm/
COPY ff-files/ ./ff-files/
COPY ff-common/ ./ff-common/
WORKDIR /src/ff-split
RUN go mod download

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux GOMAXPROCS=2 GOGC=50 go build -p 1 -o /app/server ./cmd/app

# Финальный этап
FROM alpine:latest

WORKDIR /app

# Копируем бинарный файл из этапа сборки
COPY --from=builder /app/server .

COPY ff-split/config/ ./config

# Открываем порт
EXPOSE 8085

# Запускаем приложение
CMD ["./server"] 