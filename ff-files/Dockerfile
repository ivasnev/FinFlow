# Этап сборки
FROM golang:1.24-alpine AS builder

WORKDIR /ff-files

# Копируем файлы с зависимостями
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o /ff-files/app ./cmd/

# Финальный этап
FROM alpine:latest

WORKDIR /ff-files

# Копируем бинарный файл из этапа сборки
COPY --from=builder /ff-files/ .

COPY config/ ./config

# Открываем порт
EXPOSE 8082

# Запускаем приложение
CMD ["./app"]