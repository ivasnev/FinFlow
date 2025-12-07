# FF-TVM (FinFlow Ticket Vending Machine)

## Описание

FF-TVM (FinFlow Ticket Vending Machine) - это микросервис для авторизации и управления доступом между микросервисами в системе FinFlow по механизму тикетов, аналогичный Yandex TVM. Сервис обеспечивает безопасное межсервисное взаимодействие через криптографически подписанные тикеты с использованием алгоритма ED25519.

## Функциональность

- **Генерация и валидация тикетов** - создание и проверка подписанных тикетов для межсервисного взаимодействия
- **Управление сервисами** - регистрация сервисов и управление их ключами
- **Управление правами доступа** - контроль доступа между сервисами через матрицу разрешений
- **Криптографическая защита** - использование ED25519 для подписи и проверки тикетов
- **Клиентская библиотека** - готовый клиент для использования в других сервисах
- **Middleware** - готовые middleware для проверки тикетов в HTTP запросах

## Технический стек

- **Язык:** Go 1.21+
- **Фреймворк:** Стандартная библиотека net/http
- **База данных:** PostgreSQL 12+
- **Криптография:** ED25519 (crypto/ed25519)
- **Развертывание:** Docker, Docker Compose

## Архитектура

### Структура проекта

```
ff-tvm/
├── cmd/                    # Точки входа в приложение
│   └── server/             # Основное приложение
├── config/                 # Конфигурационные файлы
│   └── config.yaml         # Конфигурация сервиса
├── internal/               # Внутренний код приложения
│   ├── api/                # API компоненты
│   │   └── handlers/       # HTTP обработчики
│   ├── config/             # Конфигурация приложения
│   ├── repository/         # Слой доступа к данным
│   │   ├── migrations/     # Миграции базы данных
│   │   └── models/         # Модели данных
│   └── service/            # Бизнес-логика
│       ├── ticket_service.go      # Сервис работы с тикетами
│       ├── key_manager.go         # Управление ключами
│       └── access_manager.go      # Управление доступом
├── pkg/                    # Публичные пакеты
│   ├── client/             # TVM клиент для других сервисов
│   ├── middleware/         # Middleware для проверки тикетов
│   └── transport/          # HTTP транспорт с автоматическим добавлением тикетов
├── docs/                   # Документация
│   ├── api.md              # API документация
│   └── tvm.postman_collection.json
├── Makefile                # Команды для разработки
├── Dockerfile              # Docker образ
└── docker-compose.yml      # Docker Compose конфигурация
```

### Архитектурные слои

1. **API Layer** - HTTP обработчики, валидация запросов
2. **Service Layer** - Бизнес-логика: генерация тикетов, управление ключами, контроль доступа
3. **Repository Layer** - Доступ к данным, работа с PostgreSQL
4. **Crypto Layer** - Криптографические операции (ED25519)

### Принцип работы

1. **Регистрация сервиса**: При регистрации сервиса генерируется пара ключей ED25519 (публичный и приватный)
2. **Настройка доступа**: Администратор настраивает матрицу доступа между сервисами
3. **Генерация тикета**: Сервис-отправитель запрашивает тикет, предоставляя свой приватный ключ
4. **Проверка тикета**: Сервис-получатель проверяет подпись тикета с помощью публичного ключа сервиса-отправителя

### Модель безопасности

- **ED25519 подписи** - все тикеты подписываются с помощью ED25519
- **TTL (Time To Live)** - тикеты имеют ограниченное время жизни (по умолчанию 24 часа)
- **Матрица доступа** - контроль доступа между сервисами через таблицу `service_access`
- **Хеширование приватных ключей** - приватные ключи хранятся в виде хешей для безопасности

## API Endpoints

### Управление сервисами

#### Создание сервиса
```http
POST /service
Content-Type: application/json
```

**Запрос:**
```json
{
  "name": "service-name"
}
```

**Ответ:**
```json
{
  "id": 1,
  "name": "service-name",
  "public_key": "base64-encoded-public-key",
  "private_key": "base64-encoded-private-key"
}
```

> **Важно**: Сохраните приватный ключ! Он понадобится для генерации тикетов.

#### Получение публичного ключа сервиса
```http
GET /service/{id}/key
```

**Ответ:**
```json
{
  "public_key": "base64-encoded-public-key"
}
```

### Управление доступом

#### Предоставление доступа
```http
POST /access/grant
Content-Type: application/json
```

**Запрос:**
```json
{
  "from": 1,
  "to": 2
}
```

**Ответ:**
```json
{
  "message": "access granted"
}
```

#### Отзыв доступа
```http
POST /access/revoke
Content-Type: application/json
```

**Запрос:**
```json
{
  "from": 1,
  "to": 2
}
```

**Ответ:**
```json
{
  "message": "access revoked"
}
```

### Генерация тикетов

#### Генерация тикета (продакшн)
```http
POST /ticket
Content-Type: application/json
```

**Запрос:**
```json
{
  "from": 1,
  "to": 2,
  "secret": "base64-encoded-private-key"
}
```

**Ответ:**
```json
{
  "from": 1,
  "to": 2,
  "ttl": 1704067200,
  "signature": "base64-encoded-signature",
  "metadata": "{}"
}
```

#### Генерация тикета для разработки
```http
POST /dev/ticket
Content-Type: application/json
```

**Запрос:**
```json
{
  "from": 1,
  "to": 2
}
```

> **Примечание**: Этот endpoint доступен только в режиме разработки (`dev.enabled: true`). В продакшене он должен быть отключен.

## Использование в других сервисах

### TVM Client

Для получения тикетов используйте TVM клиент:

```go
import "github.com/ivasnev/FinFlow/ff-tvm/pkg/client"

// Создаем клиент
tvmClient := client.NewTVMClient("http://tvm-service:8081", "your-private-key")

// Получаем публичный ключ другого сервиса
publicKey, err := tvmClient.GetPublicKey(2)

// Генерируем тикет
ticket, err := tvmClient.GenerateTicket(1, 2) // from=1, to=2
```

### HTTP Transport с автоматическим добавлением тикетов

Для автоматического добавления тикетов в заголовки HTTP запросов:

```go
import (
    "github.com/ivasnev/FinFlow/ff-tvm/pkg/client"
    "github.com/ivasnev/FinFlow/ff-tvm/pkg/transport"
    "net/http"
)

// Создаем TVM клиент
tvmClient := client.NewTVMClient("http://tvm-service:8081", "your-private-key")

// Создаем транспорт с указанием сервисов
tvmTransport := transport.NewTVMTransport(
    tvmClient,
    http.DefaultTransport,
    1, // from service ID
    2, // to service ID
)

// Создаем HTTP клиент с нашим транспортом
httpClient := &http.Client{
    Transport: tvmTransport,
}

// Теперь все запросы через этот клиент будут автоматически содержать тикет
resp, err := httpClient.Get("http://target-service/api/endpoint")
```

### Middleware для проверки тикетов

Для проверки тикетов в входящих запросах:

```go
import "github.com/ivasnev/FinFlow/ff-tvm/pkg/middleware"

// Создаем TVM клиент
tvmClient := client.NewTVMClient("http://tvm-service:8081", "")

// Создаем middleware
tvmMiddleware := middleware.NewTVMMiddleware(tvmClient)

// Используем в роутере
router.Use(tvmMiddleware.ValidateTicket())
```

## Быстрый старт

### Предварительные требования

- Go 1.21 или выше
- PostgreSQL 12 или выше
- Docker и Docker Compose (опционально)

### Настройка окружения

1. Клонировать репозиторий:
```bash
git clone https://github.com/ivasnev/FinFlow.git
cd FinFlow/FinFlowBackend/ff-tvm
```

2. Настроить конфигурацию:
```bash
cp config/config.yaml config/config.yaml.local
# Отредактировать config.yaml.local под свои настройки
```

Основные параметры конфигурации:
- `server.port` - порт сервиса (по умолчанию 8081)
- `database.url` - строка подключения к PostgreSQL
- `dev.enabled` - режим разработки (для упрощенной генерации тикетов)
- `migrations.path` - путь к миграциям

### Запуск с Docker Compose

```bash
docker-compose up -d
```

Это запустит:
- Сервис ff-tvm
- PostgreSQL (если не используется внешний)

### Локальный запуск

1. Убедитесь, что PostgreSQL запущен:
```bash
# Создать базу данных
psql -U postgres -c "CREATE DATABASE ff_tvm;"
```

2. Примените миграции:
```bash
# Используя golang-migrate
migrate -path internal/repository/migrations \
  -database "postgres://postgres:postgres@localhost:5432/ff_tvm?sslmode=disable" \
  up
```

3. Установите зависимости:
```bash
go mod download
```

4. Запустите сервис:
```bash
# Через Makefile
make run

# Или напрямую
go run cmd/server/main.go
```

Сервис будет доступен по адресу `http://localhost:8081`

## Разработка

### Сборка проекта

```bash
make build
```

Бинарный файл будет создан в `bin/ff-tvm`.

### Запуск тестов

```bash
# Все тесты
make test

# С покрытием кода
make test-cov
```

### Очистка сгенерированных файлов

```bash
make clean
```

## Конфигурация

Основные параметры конфигурации в `config/config.yaml`:

```yaml
server:
  port: "8081"

database:
  url: "postgres://postgres:postgres@localhost:5432/ff_tvm?sslmode=disable"

dev:
  enabled: false  # Режим разработки (упрощенная генерация тикетов)
  secret: "hackme"  # Секрет для dev режима

migrations:
  path: "internal/repository/migrations"
```

## Миграции базы данных

Миграции находятся в `internal/repository/migrations/`. Для применения миграций используйте `golang-migrate`:

```bash
# Применить все миграции
migrate -path internal/repository/migrations \
  -database "postgres://user:pass@localhost:5432/ff_tvm?sslmode=disable" \
  up

# Откатить последнюю миграцию
migrate -path internal/repository/migrations \
  -database "postgres://user:pass@localhost:5432/ff_tvm?sslmode=disable" \
  down 1
```

## Примеры использования

### Регистрация сервиса

```bash
curl -X POST http://localhost:8081/service \
  -H "Content-Type: application/json" \
  -d '{"name": "ff-split"}'
```

Ответ:
```json
{
  "id": 1,
  "name": "ff-split",
  "public_key": "WjA9IejKO/faO+kiyIzP3PeaxxXN6IE+eQ3Iy/0GPImp2bdWmfn2gDrM1rIJC+GOYGqTcfMwvC+YN43Vl1kPzg==",
  "private_key": "DXRZ2QMo9YKvspsua8FifxToEUFyQ5pyXqfIA49vOaB4QxvNb4MPnQjTfiSg30NctFcbVzyWW1tA6PwtZr35xw=="
}
```

### Настройка доступа между сервисами

```bash
# Предоставить доступ от ff-split (ID=1) к ff-files (ID=2)
curl -X POST http://localhost:8081/access/grant \
  -H "Content-Type: application/json" \
  -d '{"from": 1, "to": 2}'
```

### Генерация тикета

```bash
curl -X POST http://localhost:8081/ticket \
  -H "Content-Type: application/json" \
  -d '{
    "from": 1,
    "to": 2,
    "secret": "DXRZ2QMo9YKvspsua8FifxToEUFyQ5pyXqfIA49vOaB4QxvNb4MPnQjTfiSg30NctFcbVzyWW1tA6PwtZr35xw=="
  }'
```

### Использование тикета в запросе

```bash
# Получить тикет
TICKET=$(curl -s -X POST http://localhost:8081/ticket \
  -H "Content-Type: application/json" \
  -d '{"from": 1, "to": 2, "secret": "..."}' | jq -r '.ticket')

# Использовать тикет в запросе к другому сервису
curl -X GET http://localhost:8082/api/v1/files/123 \
  -H "X-TVM-Ticket: $TICKET"
```

## Безопасность

### Криптография

- **ED25519** - используется для подписи и проверки тикетов
- **Хеширование приватных ключей** - приватные ключи хранятся в виде хешей (SHA-256)
- **TTL тикетов** - тикеты имеют ограниченное время жизни (24 часа по умолчанию)

### Контроль доступа

- **Матрица доступа** - доступ между сервисами контролируется через таблицу `service_access`
- **Проверка подписи** - каждый тикет проверяется на валидность подписи
- **Проверка TTL** - тикеты с истекшим временем жизни отклоняются

### Рекомендации

1. **Храните приватные ключи в секретах** - никогда не коммитьте приватные ключи в репозиторий
2. **Используйте переменные окружения** - храните секреты в переменных окружения или системах управления секретами
3. **Отключите dev режим в продакшене** - убедитесь, что `dev.enabled: false` в продакшене
4. **Регулярно ротируйте ключи** - периодически обновляйте ключи сервисов
5. **Мониторьте доступ** - отслеживайте использование тикетов и подозрительную активность

## Интеграция с другими сервисами FinFlow

### ff-auth (ID: 1)
- Используется для валидации JWT токенов
- Получает TVM тикеты для обращения к другим сервисам

### ff-id (ID: 2)
- Использует TVM для защиты внутренних endpoints
- Получает тикеты для обращения к ff-files

### ff-files (ID: 3)
- Использует TVM для защиты всех endpoints
- Проверяет тикеты от других сервисов

### ff-split (ID: 4)
- Использует TVM для обращения к ff-id и ff-files
- Получает тикеты для межсервисного взаимодействия

## Мониторинг

Рекомендуется мониторить:
- Количество сгенерированных тикетов
- Количество проверок тикетов
- Ошибки валидации тикетов
- Время ответа API
- Использование базы данных

## Лицензия

MIT
