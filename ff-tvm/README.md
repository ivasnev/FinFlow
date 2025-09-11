# FF-TVM (FinFlow Ticket Vending Machine)

Микросервис для авторизации микросервисов в системе по механизму тикетов, аналогичный Yandex TVM.

## Функциональность

- Генерация и валидация тикетов для межсервисного взаимодействия
- Управление сервисами и их ключами
- Управление правами доступа между сервисами
- Middleware для проверки тикетов в других сервисах

## Требования

- Go 1.21 или выше
- PostgreSQL 12 или выше

## Установка

1. Клонируйте репозиторий:
```bash
git clone https://github.com/ivasnev/FinFlow.git
cd FinFlow/ff-tvm
```

2. Установите зависимости:
```bash
go mod download
```

3. Создайте базу данных и примените миграции:
```bash
psql -U postgres -d finflow -f migrations/001_init.sql
```

4. Запустите сервис:
```bash
export DATABASE_URL="postgres://user:password@localhost:5432/finflow?sslmode=disable"
go run cmd/server/main.go
```

## API

### Создание сервиса
```http
POST /service
Content-Type: application/json

{
    "name": "service-name",
    "public_key": "base64-encoded-public-key"
}
```

### Получение публичного ключа сервиса
```http
GET /service/{id}/pub_key
```

### Создание тикета
```http
POST /ticket
Content-Type: application/json

{
    "from": 1,
    "to": 2
}
```

## Middleware

Для использования в других сервисах добавьте middleware:

```go
import "github.com/ivasnev/FinFlow/ff-tvm/pkg/middleware"

client := client.NewTVMClient("http://tvm-service:8080")
tvmMiddleware := middleware.NewTVMMiddleware(client)

router.Use(tvmMiddleware.ValidateTicket())
```

## Безопасность

- Все тикеты подписываются с помощью ED25519
- Ключи хранятся в базе данных в зашифрованном виде
- Доступ между сервисами контролируется через таблицу `service_access`

## Лицензия

MIT 