# FinFlow Identity Service (ff-auth)

## Описание
FinFlow Identity Service (ff-auth) - это сервис аутентификации и управления пользователями для экосистемы FinFlow. Сервис предоставляет API для регистрации, аутентификации пользователей, управления сессиями и ролями.

## Функциональность

- **Регистрация и аутентификация пользователей**
- **Управление сессиями**
- **Управление профилем пользователя**
- **Контроль доступа на основе ролей**
- **Отслеживание истории входов**
- **Управление устройствами**

## Технический стек

- **Язык:** Go
- **Фреймворк:** Gin (HTTP-сервер)
- **База данных:** PostgreSQL
- **ORM:** GORM
- **Аутентификация:** JWT
- **Документация API:** Swagger

## Архитектура

### Структура проекта

```
ff-auth/
├── cmd/                # Точки входа в приложение
│   └── app/            # Основное приложение
├── config/             # Конфигурационные файлы
│   └── config.yaml     # Конфигурация сервиса
├── internal/           # Внутренний код приложения
│   ├── api/            # API компоненты
│   │   ├── dto/        # Объекты передачи данных
│   │   ├── handler/    # Обработчики HTTP-запросов
│   │   └── middleware/ # Промежуточные обработчики
│   ├── app/            # Ядро приложения
│   ├── adapters/       # Адаптеры внешних сервисов
│   │   └── ffid/       # Адаптер для ff-id сервиса
│   ├── common/         # Общие компоненты
│   │   └── config/     # Конфигурация приложения
│   ├── container/      # Контейнер зависимостей (DI)
│   ├── models/         # Модели данных
│   ├── repository/     # Слой доступа к данным
│   │   └── postgres/   # Реализация на PostgreSQL
│   └── service/        # Бизнес-логика
│       ├── auth/       # Сервис аутентификации
│       ├── session/    # Сервис сессий
│       └── user/       # Сервис пользователей
├── pkg/                # Публичные пакеты
│   ├── api/            # API спецификация
│   └── auth/           # Публичные пакеты для аутентификации
├── tests/              # Тесты
│   ├── base_suite.go   # Базовый набор тестов
│   ├── login_test.go   # Тесты входа
│   ├── register_test.go # Тесты регистрации
│   └── migrations.sql  # SQL для тестов
├── Makefile            # Команды для разработки
├── Dockerfile          # Docker образ
└── docker-compose.yml  # Docker Compose конфигурация
```

### Архитектурные слои

1. **API Layer** - HTTP обработчики, валидация запросов, обработка ответов
2. **Service Layer** - Бизнес-логика: аутентификация, управление сессиями, управление пользователями
3. **Repository Layer** - Доступ к данным, работа с PostgreSQL
4. **Adapter Layer** - Интеграция с внешними сервисами (ff-id, ff-tvm)

### Взаимодействие с другими сервисами

- **ff-id** - создание профиля пользователя при регистрации
- **ff-tvm** - получение TVM тикетов для межсервисного взаимодействия
- **ff-split** - валидация JWT токенов для доступа к API

## API Endpoints

### Аутентификация

#### Регистрация нового пользователя
```
POST /api/v1/auth/register
```
Регистрирует нового пользователя в системе.

**Запрос:**
```json
{
  "email": "user@example.com",
  "password": "StrongPassword123",
  "nickname": "user123",
  "name": "John Doe",
  "phone": "+71234567890"
}
```

**Ответ:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2023-01-01T12:00:00Z",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "nickname": "user123",
    "roles": ["user"],
    "created_at": "2023-01-01T10:00:00Z",
    "updated_at": "2023-01-01T10:00:00Z"
  }
}
```

#### Вход в систему
```
POST /api/v1/auth/login
```
Аутентифицирует пользователя и создает новую сессию.

**Запрос:**
```json
{
  "login": "user@example.com",
  "password": "StrongPassword123"
}
```

**Ответ:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2023-01-01T12:00:00Z",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "nickname": "user123",
    "roles": ["user"],
    "created_at": "2023-01-01T10:00:00Z",
    "updated_at": "2023-01-01T10:00:00Z"
  }
}
```

#### Обновление токена
```
POST /api/v1/auth/refresh
```
Обновляет access-токен с помощью refresh-токена.

**Запрос:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Ответ:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2023-01-01T14:00:00Z",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "nickname": "user123",
    "roles": ["user"],
    "created_at": "2023-01-01T10:00:00Z",
    "updated_at": "2023-01-01T10:00:00Z"
  }
}
```

#### Выход из системы
```
POST /api/v1/auth/logout
```
Завершает текущую сессию пользователя.

**Запрос:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Ответ:**
```json
{
  "message": "successfully logged out"
}
```

### Пользователи

#### Получение информации о пользователе по никнейму
```
GET /api/v1/users/:nickname
```
Возвращает информацию о пользователе по его никнейму.

**Ответ:**
```json
{
  "id": 1,
  "email": "user@example.com",
  "nickname": "user123",
  "name": "John Doe",
  "phone": "+71234567890",
  "created_at": "2023-01-01T10:00:00Z",
  "updated_at": "2023-01-01T10:00:00Z"
}
```

#### Обновление профиля пользователя
```
PATCH /api/v1/users/me
```
Обновляет информацию о текущем пользователе.

**Запрос:**
```json
{
  "nickname": "updated_nickname",
  "name": "Updated Name",
  "phone": "+79876543210"
}
```

**Ответ:**
```json
{
  "id": 1,
  "email": "user@example.com",
  "nickname": "updated_nickname",
  "name": "Updated Name",
  "phone": "+79876543210",
  "created_at": "2023-01-01T10:00:00Z",
  "updated_at": "2023-01-01T11:00:00Z"
}
```

### Сессии

#### Получение активных сессий
```
GET /api/v1/sessions
```
Возвращает список активных сессий текущего пользователя.

**Ответ:**
```json
[
  {
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "device": "Mozilla/5.0 Chrome/92.0.4515.131",
    "ip_address": "192.168.1.1",
    "created_at": "2023-01-01T10:00:00Z",
    "expires_at": "2023-01-08T10:00:00Z"
  },
  {
    "id": "b2c3d4e5-f6a7-8901-bcde-f12345678901",
    "device": "Mozilla/5.0 Safari/605.1.15",
    "ip_address": "192.168.1.2",
    "created_at": "2023-01-02T11:00:00Z",
    "expires_at": "2023-01-09T11:00:00Z"
  }
]
```

#### Завершение сессии
```
DELETE /api/v1/sessions/:id
```
Завершает указанную сессию текущего пользователя.

**Ответ:**
```json
{
  "message": "session terminated successfully"
}
```

### История входов

#### Получение истории входов
```
GET /api/v1/login-history
```
Возвращает историю входов текущего пользователя.

**Параметры запроса:**
- `limit` - количество записей на странице (по умолчанию 10)
- `offset` - смещение для пагинации (по умолчанию 0)

**Ответ:**
```json
[
  {
    "id": 1,
    "ip_address": "192.168.1.1",
    "user_agent": "Mozilla/5.0 Chrome/92.0.4515.131",
    "created_at": "2023-01-01T10:00:00Z"
  },
  {
    "id": 2,
    "ip_address": "192.168.1.2",
    "user_agent": "Mozilla/5.0 Safari/605.1.15",
    "created_at": "2023-01-02T11:00:00Z"
  }
]
```

## Аутентификация

Для доступа к защищенным ресурсам требуется аутентификация с помощью JWT-токена. Токен должен быть передан в заголовке Authorization в формате:

```
Authorization: Bearer {access_token}
```

## Обработка ошибок

API возвращает стандартные HTTP-коды состояния:

- `200 OK` - запрос выполнен успешно
- `201 Created` - ресурс успешно создан
- `400 Bad Request` - неверные данные запроса
- `401 Unauthorized` - отсутствие или недействительность токена аутентификации
- `403 Forbidden` - доступ запрещен (недостаточно прав)
- `404 Not Found` - ресурс не найден
- `500 Internal Server Error` - внутренняя ошибка сервера

Ответ с ошибкой содержит сообщение об ошибке в формате:

```json
{
  "error": "описание ошибки"
}
```

## Быстрый старт

### Предварительные требования

- Go 1.19 или выше
- PostgreSQL 12 или выше
- Docker и Docker Compose (опционально)

### Настройка окружения

1. Клонировать репозиторий:
```bash
git clone https://github.com/ivasnev/FinFlow.git
cd FinFlow/FinFlowBackend/ff-auth
```

2. Настроить конфигурацию:
```bash
cp config/config.yaml config/config.yaml.local
# Отредактировать config.yaml.local под свои настройки
```

Основные параметры конфигурации:
- `server.port` - порт сервиса (по умолчанию 8084)
- `postgres.*` - параметры подключения к PostgreSQL
- `auth.*` - настройки JWT токенов (секрет, время жизни)
- `id_client.*` - настройки интеграции с ff-id
- `tvm.*` - настройки TVM для межсервисной авторизации

### Запуск с Docker Compose

```bash
docker-compose up -d
```

Это запустит:
- Сервис ff-auth
- PostgreSQL (если не используется внешний)

### Локальный запуск

1. Убедитесь, что PostgreSQL запущен:
```bash
# Создать базу данных
psql -U postgres -c "CREATE DATABASE ff_auth;"
```

2. Примените миграции:
```bash
# Используя golang-migrate
migrate -path internal/repository/migrations \
  -database "postgres://postgres:postgres@localhost:5432/ff_auth?sslmode=disable" \
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
go run cmd/app/main.go
```

Сервис будет доступен по адресу `http://localhost:8084`

## Разработка

### Сборка проекта
```
go build -o ./bin/ff-auth ./cmd/app/
```

### Запуск тестов
```
go test ./...
```

### Запуск линтера
```
golangci-lint run
``` 