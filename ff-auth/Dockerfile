# Этап сборки
FROM golang:1.23-alpine AS builder

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
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/app

# Финальный этап
FROM alpine:latest

WORKDIR /app

# Копируем бинарный файл из этапа сборки
COPY --from=builder /app/server .

COPY config/ ./config

# Открываем порт
EXPOSE 8084

# Запускаем приложение
CMD ["./server"]