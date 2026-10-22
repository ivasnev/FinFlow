# Этап сборки
FROM golang:1.27.3-alpine AS builder

# Устанавливаем git для скачивания Go модулей
RUN apk add --no-cache git

# Сборка из корня репозитория включает локальную TVM-проверку.
WORKDIR /src
COPY ff-files/ ./ff-files/
COPY ff-tvm/ ./ff-tvm/
COPY ff-common/ ./ff-common/
WORKDIR /src/ff-files
RUN go mod download

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux GOMAXPROCS=2 GOGC=50 go build -p 1 -o /app/server ./cmd/server

# Финальный этап
FROM alpine:latest

WORKDIR /app

# Копируем бинарный файл из этапа сборки
COPY --from=builder /app/server .

COPY ff-files/config/ ./config

# Открываем порт
EXPOSE 8082

# Запускаем приложение
CMD ["./server"]