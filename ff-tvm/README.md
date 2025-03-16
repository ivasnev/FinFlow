# FF-TVM Service (Trusted Verification Module)

Сервис для управления авторизацией между микросервисами в экосистеме FinFlow. Реализует механизм тикетов для безопасного взаимодействия сервисов.

## Функциональность

- Регистрация микросервисов
- Управление доступами между сервисами
- Выдача и валидация тикетов
- Управление ключами (RSA)
- Ротация ключей

## Технологии

- Go 1.21+
- PostgreSQL
- Redis
- Gin Web Framework
- GORM
- JWT (RS256)

## Зависимости

- PostgreSQL 14+
- Redis 6+

## Установка и запуск

1. Клонируйте репозиторий:
```bash
git clone https://github.com/ivasnev/FinFlow.git
cd FinFlow/ff-tvm
```

2. Установите зависимости:
```bash
go mod download
```

3. Настройте конфигурацию в файле `config.yaml`

4. Запустите PostgreSQL и Redis

5. Запустите сервис:
```bash
go run cmd/app/main.go
```

## API Endpoints

### Управление сервисами

- `POST /tvm/register` - Регистрация нового сервиса
  ```json
  {
    "name": "service-name",
    "description": "Service description"
  }
  ```

- `POST /tvm/access/grant` - Предоставление доступа
  ```json
  {
    "source_service_id": 1,
    "target_service_id": 2
  }
  ```

- `POST /tvm/access/revoke` - Отзыв доступа
  ```json
  {
    "source_service_id": 1,
    "target_service_id": 2
  }
  ```

### Управление тикетами

- `POST /tvm/ticket` - Получение тикета
  ```json
  {
    "source_service_id": 1,
    "target_service_id": 2
  }
  ```

- `POST /tvm/validate` - Проверка тикета
  ```json
  {
    "ticket": "jwt-token"
  }
  ```

### Управление ключами

- `GET /tvm/public-key/:service_id` - Получение публичного ключа
- `POST /tvm/rotate-keys/:service_id` - Ротация ключей

## Структура проекта

```
ff-tvm/
├── cmd/
│   └── app/
│       └── main.go
├── internal/
│   ├── config/
│   ├── handler/
│   ├── models/
│   ├── repository/
│   └── service/
├── pkg/
│   └── crypto/
├── go.mod
├── go.sum
└── README.md
```

## Безопасность

- Использование RSA для подписи тикетов
- Проверка прав доступа при выдаче тикетов
- Кеширование тикетов в Redis
- Ротация ключей
- Валидация всех входящих запросов

## Интеграция

Для интеграции с FF-TVM другие сервисы должны:

1. Зарегистрироваться в TVM и получить ID
2. Получить тикет для доступа к целевому сервису
3. Использовать тикет в заголовке запроса
4. Целевой сервис должен проверять тикет через TVM 