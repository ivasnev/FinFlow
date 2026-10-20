# Сборка из корня репозитория использует согласованные локальные версии auth, id и tvm.
FROM golang:1.24-alpine AS builder
RUN apk add --no-cache git
WORKDIR /src
COPY ff-auth/ ./ff-auth/
COPY ff-id/ ./ff-id/
COPY ff-tvm/ ./ff-tvm/
WORKDIR /src/ff-auth
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/app

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .
COPY ff-auth/config/ ./config
EXPOSE 8084
CMD ["./server"]
