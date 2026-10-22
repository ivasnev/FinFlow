# Этап сборки
FROM golang:1.27.3-alpine AS builder

# Устанавливаем git для скачивания Go модулей
RUN apk add --no-cache git

WORKDIR /app

# Копируем файлы с зависимостями
COPY go.mod go.sum ./

RUN sed -i '/replace (/,/)/d' go.mod
RUN sed -i '/replace .*=>/d' go.mod

# Скачиваем зависимости
RUN go mod download all

# Копируем исходный код
COPY internal/ ./internal
COPY cmd/ ./cmd
COPY pkg/ ./pkg
COPY config/ ./config

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux GOMAXPROCS=2 GOGC=50 go build -p 1 -o /app/server ./cmd/server

# Финальный этап
FROM alpine:latest

WORKDIR /app

# Копируем бинарный файл из этапа сборки
COPY --from=builder /app/server .

# Копируем миграции
COPY internal/repository/migrations ./migrations
COPY config/ ./config

# Открываем порт
EXPOSE 8081

# Запускаем приложение
CMD ["./server"] 