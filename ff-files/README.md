# FinFlow Files Service (ff-files)

## Описание

FinFlow Files Service (ff-files) - это сервис для управления файлами в экосистеме FinFlow. Сервис предоставляет API для загрузки, хранения, получения и удаления файлов с использованием MinIO в качестве объектного хранилища. Сервис используется для хранения чеков, аватаров пользователей и других файлов, связанных с финансовыми операциями.

## Функциональность

- **Загрузка файлов** - одиночная и массовая загрузка файлов в MinIO
- **Получение файлов** - получение файлов по идентификатору с генерацией временных ссылок
- **Удаление файлов** - удаление одного или нескольких файлов
- **Метаданные файлов** - получение информации о файлах (размер, тип, дата загрузки)
- **Временные ссылки** - генерация подписанных временных ссылок для безопасного доступа к файлам
- **Поддержка различных форматов** - изображения, документы и другие типы файлов

## Технический стек

- **Язык:** Go 1.19+
- **Фреймворк:** Gin (HTTP-сервер)
- **База данных:** PostgreSQL 13+ (для метаданных)
- **Объектное хранилище:** MinIO (S3-совместимое)
- **Кэш:** Redis (опционально)
- **Аутентификация:** TVM (Ticket-based Verification Module) для межсервисного взаимодействия
- **Документация API:** OpenAPI 3.0

## Архитектура

### Структура проекта

```
ff-files/
├── cmd/                    # Точки входа в приложение
│   └── server/             # Основное приложение
├── config/                 # Конфигурационные файлы
│   └── config.yaml         # Конфигурация сервиса
├── internal/               # Внутренний код приложения
│   ├── api/                # API компоненты
│   │   ├── handler/        # Обработчики HTTP-запросов
│   │   └── middleware/     # Промежуточные обработчики (CORS)
│   ├── app/                # Ядро приложения
│   ├── common/             # Общие компоненты
│   │   ├── config/         # Конфигурация приложения
│   │   ├── dto/            # Data Transfer Objects
│   │   ├── errors/         # Обработка ошибок
│   │   └── responses/      # Стандартные ответы
│   ├── container/          # Контейнер зависимостей (DI)
│   └── service/            # Бизнес-логика
│       └── minio/          # Сервис работы с MinIO
├── pkg/                    # Публичные пакеты
│   └── api/                # Сгенерированные API типы из OpenAPI
├── Makefile                # Команды для разработки
├── Dockerfile              # Docker образ
└── docker-compose.yml      # Docker Compose конфигурация
```

### Архитектурные слои

1. **API Layer** - HTTP обработчики, валидация запросов, обработка multipart/form-data
2. **Service Layer** - Бизнес-логика работы с файлами, взаимодействие с MinIO
3. **Storage Layer** - MinIO для хранения файлов, PostgreSQL для метаданных (если используется)

### Взаимодействие с другими сервисами

- **ff-tvm** - получение TVM тикетов для межсервисной авторизации
- **ff-split** - загрузка чеков для транзакций
- **ff-id** - загрузка аватаров пользователей

## API Endpoints

### Загрузка файлов

#### Загрузка одного файла
```
POST /api/v1/files
Content-Type: multipart/form-data
X-TVM-Ticket: {tvm_ticket}
```

**Параметры:**
- `file` (обязательный) - файл для загрузки
- `metadata` (опциональный) - метаданные в формате JSON

**Ответ:**
```json
{
  "status": 200,
  "message": "File uploaded successfully",
  "data": {
    "object_id": "26296f14-3f08-4f79-b453-09ebc4eac98d",
    "link": "http://localhost:9000/ff-files/26296f14-3f08-4f79-b453-09ebc4eac98d"
  }
}
```

#### Загрузка нескольких файлов
```
POST /api/v1/files/many
Content-Type: multipart/form-data
X-TVM-Ticket: {tvm_ticket}
```

**Параметры:**
- `files` (обязательный) - массив файлов для загрузки

**Ответ:**
```json
{
  "status": 200,
  "message": "Files uploaded successfully",
  "data": [
    {
      "object_id": "26296f14-3f08-4f79-b453-09ebc4eac98d",
      "link": "http://localhost:9000/ff-files/26296f14-3f08-4f79-b453-09ebc4eac98d"
    },
    {
      "object_id": "37397e25-4g19-5g8a-c564-10fcd5fbd09e",
      "link": "http://localhost:9000/ff-files/37397e25-4g19-5g8a-c564-10fcd5fbd09e"
    }
  ]
}
```

### Получение файлов

#### Получение одного файла
```
GET /api/v1/files/{file_id}
X-TVM-Ticket: {tvm_ticket}
```

**Ответ:**
```json
{
  "status": 200,
  "message": "File received successfully",
  "data": "http://localhost:9000/ff-files/26296f14-3f08-4f79-b453-09ebc4eac98d"
}
```

#### Получение нескольких файлов
```
GET /api/v1/files/many
Content-Type: application/json
X-TVM-Ticket: {tvm_ticket}
```

**Запрос:**
```json
{
  "fileIDs": [
    "26296f14-3f08-4f79-b453-09ebc4eac98d",
    "37397e25-4g19-5g8a-c564-10fcd5fbd09e"
  ]
}
```

**Ответ:**
```json
{
  "status": 200,
  "message": "Files received successfully",
  "data": [
    "http://localhost:9000/ff-files/26296f14-3f08-4f79-b453-09ebc4eac98d",
    "http://localhost:9000/ff-files/37397e25-4g19-5g8a-c564-10fcd5fbd09e"
  ]
}
```

#### Получение метаданных файла
```
GET /api/v1/files/{file_id}/metadata
X-TVM-Ticket: {tvm_ticket}
```

**Ответ:**
```json
{
  "status": 200,
  "message": "File metadata retrieved successfully",
  "data": {
    "file_id": "26296f14-3f08-4f79-b453-09ebc4eac98d",
    "filename": "receipt.jpg",
    "size": 1048576,
    "content_type": "image/jpeg",
    "upload_date": "2024-01-01T10:00:00Z",
    "owner_id": "user123",
    "metadata": {
      "description": "Чек из магазина",
      "tags": ["чек", "покупка"]
    }
  }
}
```

### Генерация временной ссылки

```
POST /api/v1/files/{file_id}/temporary-url?expires_in=3600
X-TVM-Ticket: {tvm_ticket}
```

**Параметры запроса:**
- `expires_in` (опциональный) - время жизни ссылки в секундах (по умолчанию 1 час, максимум 24 часа)

**Ответ:**
```json
{
  "status": 200,
  "message": "Temporary URL generated successfully",
  "data": {
    "url": "http://localhost:9000/ff-files/26296f14-3f08-4f79-b453-09ebc4eac98d?X-Amz-Algorithm=...",
    "expires_at": "2024-01-01T13:00:00Z"
  }
}
```

### Удаление файлов

#### Удаление одного файла
```
DELETE /api/v1/files/{file_id}
X-TVM-Ticket: {tvm_ticket}
```

**Ответ:**
```json
{
  "status": 200,
  "message": "File deleted successfully"
}
```

#### Удаление нескольких файлов
```
DELETE /api/v1/files/many
Content-Type: application/json
X-TVM-Ticket: {tvm_ticket}
```

**Запрос:**
```json
{
  "fileIDs": [
    "26296f14-3f08-4f79-b453-09ebc4eac98d",
    "37397e25-4g19-5g8a-c564-10fcd5fbd09e"
  ]
}
```

**Ответ:**
```json
{
  "status": 200,
  "message": "Files deleted successfully"
}
```

## Аутентификация

Все endpoints требуют аутентификации через TVM (Ticket-based Verification Module) для межсервисного взаимодействия:

```
X-TVM-Ticket: {tvm_ticket}
```

TVM тикет должен быть получен от сервиса ff-tvm перед обращением к ff-files.

## Поддерживаемые форматы файлов

- **Изображения:** JPG, JPEG, PNG, GIF, BMP, WEBP
- **Документы:** PDF, DOC, DOCX, XLS, XLSX, TXT
- **Максимальный размер файла:** 100MB (настраивается)

## Обработка ошибок

API возвращает стандартные HTTP-коды состояния:

- `200 OK` - запрос выполнен успешно
- `400 Bad Request` - неверные данные запроса (файл не передан, неверный формат)
- `404 Not Found` - файл не найден
- `500 Internal Server Error` - внутренняя ошибка сервера

Ответ с ошибкой содержит сообщение об ошибке в формате:

```json
{
  "error": "описание ошибки",
  "code": 400
}
```

## Быстрый старт

### Предварительные требования

- Go 1.19 или выше
- MinIO (локально или удаленно)
- PostgreSQL 13+ (опционально, для метаданных)
- Redis (опционально, для кэширования)
- Docker и Docker Compose (опционально)

### Настройка окружения

1. Клонировать репозиторий:
```bash
git clone https://github.com/ivasnev/FinFlow.git
cd FinFlow/FinFlowBackend/ff-files
```

2. Настроить конфигурацию:
```bash
cp config/config.yaml config/config.yaml.local
# Отредактировать config.yaml.local под свои настройки
```

Основные параметры конфигурации:
- `server.port` - порт сервиса (по умолчанию 8082)
- `minio.*` - параметры подключения к MinIO
- `postgres.*` - параметры подключения к PostgreSQL (если используется)
- `redis.*` - параметры подключения к Redis (если используется)
- `tvm.*` - настройки TVM для межсервисной авторизации

### Запуск MinIO

#### Локально через Docker:
```bash
docker run -d \
  -p 9000:9000 \
  -p 9001:9001 \
  --name minio \
  -e "MINIO_ROOT_USER=minioadmin" \
  -e "MINIO_ROOT_PASSWORD=minioadmin" \
  minio/minio server /data --console-address ":9001"
```

#### Или через Docker Compose:
```bash
docker-compose up -d minio
```

### Запуск с Docker Compose

```bash
docker-compose up -d
```

Это запустит:
- Сервис ff-files
- MinIO
- PostgreSQL (если не используется внешний)
- Redis (если не используется внешний)

### Локальный запуск

1. Убедитесь, что MinIO запущен и доступен:
```bash
# Проверка доступности MinIO
curl http://localhost:9000/minio/health/live
```

2. Создайте bucket в MinIO (если еще не создан):
```bash
# Через MinIO Client (mc)
mc alias set myminio http://localhost:9000 minioadmin minioadmin
mc mb myminio/ff-files
```

Или через веб-консоль MinIO: http://localhost:9001

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

Сервис будет доступен по адресу `http://localhost:8082`

## Разработка

### Генерация API кода из OpenAPI спецификации

```bash
make generate
```

Это сгенерирует типы, клиент и сервер из `pkg/api/openapi.yaml`.

### Сборка проекта

```bash
make build
```

Бинарный файл будет создан в `bin/ff-files`.

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
  port: 8082

postgres:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: ff_files

minio:
  endpoint: localhost:9000
  internal_endpoint: http://localhost:9000
  access_key: minioadmin
  secret_key: minioadmin
  bucket: ff-files
  use_ssl: false
  root_user: minioadmin
  root_password: minioadmin
  file_time_expiration: 300  # Время жизни временных ссылок (в секундах)

redis:
  host: localhost
  port: 6379
  password: ""

tvm:
  base_url: http://localhost:8081
  service_id: 3
  service_secret: "..."  # Секрет для TVM

migrations:
  path: migrations
```

## Работа с MinIO

### Структура хранения

Файлы хранятся в MinIO bucket с использованием UUID в качестве имени объекта:
- Bucket: `ff-files` (настраивается)
- Имя объекта: UUID файла (например, `26296f14-3f08-4f79-b453-09ebc4eac98d`)

### Временные ссылки

Сервис генерирует подписанные временные ссылки (presigned URLs) для безопасного доступа к файлам:
- По умолчанию ссылки действительны 1 час
- Максимальное время жизни: 24 часа
- Ссылки содержат подпись MinIO и не требуют дополнительной аутентификации

### Метаданные файлов

Метаданные файлов могут храниться:
- В объектах MinIO (как метаданные объекта)
- В PostgreSQL (если используется)
- В Redis (для кэширования)

## Безопасность

- Все endpoints защищены TVM тикетами
- Временные ссылки имеют ограниченное время жизни
- Файлы хранятся с уникальными UUID именами
- Поддержка SSL/TLS для MinIO (настраивается)
- Валидация типов и размеров файлов
- CORS настроен для безопасного взаимодействия

## Производительность

- Параллельная загрузка нескольких файлов
- Кэширование метаданных в Redis (если используется)
- Эффективное использование MinIO для хранения больших файлов
- Генерация временных ссылок без обращения к базе данных

## Мониторинг

Рекомендуется мониторить:
- Размер bucket в MinIO
- Количество загруженных файлов
- Время ответа API
- Ошибки при работе с MinIO

## Интеграция с другими сервисами

### Загрузка чеков для транзакций (ff-split)

```bash
# 1. Получить TVM тикет от ff-tvm
# 2. Загрузить файл чека
curl -X POST http://localhost:8082/api/v1/files \
  -H "X-TVM-Ticket: {ticket}" \
  -F "file=@receipt.jpg"

# 3. Использовать object_id в транзакции
```

### Загрузка аватаров (ff-id)

Аналогично, ff-id использует ff-files для хранения аватаров пользователей.

## Лицензия

MIT
