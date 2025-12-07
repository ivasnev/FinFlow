# FinFlow Identity Service (ff-id)

## Описание

FinFlow Identity Service (ff-id) - это сервис управления профилями пользователей и системой друзей для экосистемы FinFlow. Сервис предоставляет API для управления пользовательскими профилями, работы с системой друзей и межсервисного взаимодействия для синхронизации данных пользователей.

## Функциональность

- **Управление профилями пользователей** - получение и обновление информации о пользователях
- **Система друзей** - отправка заявок в друзья, принятие/отклонение заявок, просмотр списка друзей
- **Управление аватарами** - загрузка и управление аватарами пользователей
- **Межсервисное взаимодействие** - внутренние API для регистрации пользователей и получения данных по ID
- **Кэширование данных** - использование Redis для кэширования профилей пользователей

## Технический стек

- **Язык:** Go 1.19+
- **Фреймворк:** Gin (HTTP-сервер)
- **База данных:** PostgreSQL 13+
- **Кэш:** Redis
- **ORM:** GORM
- **Аутентификация:** JWT (валидация через ff-auth), TVM (межсервисное взаимодействие)
- **Документация API:** OpenAPI 3.0

## Архитектура

### Структура проекта

```
ff-id/
├── cmd/                    # Точки входа в приложение
│   └── app/                # Основное приложение
├── config/                 # Конфигурационные файлы
│   └── config.yaml         # Конфигурация сервиса
├── internal/               # Внутренний код приложения
│   ├── api/                # API компоненты
│   │   ├── handler/        # Обработчики HTTP-запросов
│   │   └── middleware/      # Промежуточные обработчики (CORS)
│   ├── app/                # Ядро приложения
│   ├── common/             # Общие компоненты
│   │   ├── config/         # Конфигурация приложения
│   │   └── validator/      # Валидаторы данных
│   ├── container/          # Контейнер зависимостей (DI)
│   ├── models/             # Модели данных
│   ├── repository/         # Слой доступа к данным
│   │   ├── postgres/       # Реализация на PostgreSQL
│   │   │   ├── user/       # Репозиторий пользователей
│   │   │   ├── friend/     # Репозиторий друзей
│   │   │   └── avatar/     # Репозиторий аватаров
│   │   └── migrations/     # Миграции базы данных
│   └── service/            # Бизнес-логика
│       ├── user/           # Сервис пользователей
│       └── friend/          # Сервис друзей
├── pkg/                    # Публичные пакеты
│   └── api/                # Сгенерированные API типы из OpenAPI
├── tests/                  # Тесты
│   ├── base_suite.go       # Базовый набор тестов
│   ├── user_test.go        # Тесты пользователей
│   ├── friend_test.go      # Тесты друзей
│   └── migrations.sql      # SQL для тестов
├── Makefile                # Команды для разработки
├── Dockerfile              # Docker образ
└── docker-compose.yml      # Docker Compose конфигурация
```

### Архитектурные слои

1. **API Layer** - HTTP обработчики, валидация запросов
2. **Service Layer** - Бизнес-логика, работа с друзьями и профилями
3. **Repository Layer** - Доступ к данным, работа с PostgreSQL
4. **External Services** - Интеграция с ff-auth (валидация JWT), ff-files (аватары), ff-tvm (межсервисная авторизация)

### Взаимодействие с другими сервисами

- **ff-auth** - валидация JWT токенов, получение публичных ключей для проверки токенов
- **ff-files** - загрузка и получение аватаров пользователей
- **ff-tvm** - получение TVM тикетов для межсервисного взаимодействия
- **ff-split** - предоставление данных о пользователях по ID

## API Endpoints

### Пользователи

#### Получение информации о пользователе по никнейму
```
GET /api/v1/users/{nickname}
```
Публичный endpoint для получения информации о пользователе.

**Ответ:**
```json
{
  "id": 1,
  "email": "user@example.com",
  "nickname": "johndoe",
  "name": "John Doe",
  "phone": "+71234567890",
  "birthdate": 631152000,
  "avatar_id": "26296f14-3f08-4f79-b453-09ebc4eac98d",
  "created_at": "2023-01-01T10:00:00Z",
  "updated_at": "2023-01-01T10:00:00Z"
}
```

#### Обновление профиля текущего пользователя
```
PATCH /api/v1/users/me
```
Обновляет информацию о текущем пользователе. Требует JWT аутентификации.

**Запрос:**
```json
{
  "nickname": "new_nickname",
  "name": "Updated Name",
  "phone": "+79876543210",
  "birthdate": 631152000
}
```

#### Регистрация нового пользователя (внутренний)
```
POST /api/v1/users/register
```
Внутренний endpoint для регистрации пользователя из ff-auth. Требует TVM аутентификации.

**Запрос:**
```json
{
  "user_id": 1,
  "email": "user@example.com",
  "nickname": "johndoe",
  "name": "John Doe",
  "phone": "+71234567890"
}
```

#### Получение пользователей по ID (внутренний)
```
GET /api/v1/users/ids?ids=1,2,3
```
Внутренний endpoint для получения данных о пользователях по их ID. Требует TVM аутентификации.

### Друзья

#### Получение списка друзей
```
GET /api/v1/users/{nickname}/friends?page=1&page_size=20&name_filter=John
```
Публичный endpoint для получения списка друзей пользователя с пагинацией.

**Ответ:**
```json
{
  "friends": [
    {
      "id": 2,
      "nickname": "janedoe",
      "name": "Jane Doe",
      "avatar_id": "37397e25-4g19-5g8a-c564-10fcd5fbd09e",
      "status": "accepted"
    }
  ],
  "total": 10,
  "page": 1,
  "page_size": 20
}
```

#### Отправка заявки в друзья
```
POST /api/v1/friends/request
```
Отправляет заявку в друзья другому пользователю. Требует JWT аутентификации.

**Запрос:**
```json
{
  "friend_nickname": "janedoe"
}
```

#### Принятие заявки в друзья
```
POST /api/v1/friends/accept
```
Принимает заявку в друзья. Требует JWT аутентификации.

**Запрос:**
```json
{
  "friend_id": 2
}
```

#### Отклонение заявки в друзья
```
POST /api/v1/friends/reject
```
Отклоняет заявку в друзья. Требует JWT аутентификации.

**Запрос:**
```json
{
  "friend_id": 2
}
```

#### Удаление из друзей
```
DELETE /api/v1/friends/{friend_id}
```
Удаляет пользователя из списка друзей. Требует JWT аутентификации.

#### Получение заявок в друзья
```
GET /api/v1/friends/requests?status=pending
```
Получает список заявок в друзья (входящих и исходящих). Требует JWT аутентификации.

**Параметры запроса:**
- `status` - фильтр по статусу: `pending`, `accepted`, `rejected`, `blocked`

## Аутентификация

### Для пользователей (JWT)

Защищенные endpoints требуют аутентификации через JWT токен, полученный от ff-auth:

```
Authorization: Bearer {access_token}
```

Сервис автоматически валидирует токен через ff-auth, получая публичные ключи с заданным интервалом обновления.

### Для межсервисного взаимодействия (TVM)

Внутренние endpoints требуют TVM аутентификации:

```
X-TVM-Ticket: {tvm_ticket}
```

## Статусы дружбы

- `pending` - заявка в друзья ожидает подтверждения
- `accepted` - дружба подтверждена
- `rejected` - заявка отклонена
- `blocked` - пользователь заблокирован

## Обработка ошибок

API возвращает стандартные HTTP-коды состояния:

- `200 OK` - запрос выполнен успешно
- `201 Created` - ресурс успешно создан
- `400 Bad Request` - неверные данные запроса
- `401 Unauthorized` - отсутствие или недействительность токена аутентификации
- `403 Forbidden` - доступ запрещен
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
- PostgreSQL 13 или выше
- Redis 6 или выше
- Docker и Docker Compose (опционально)

### Настройка окружения

1. Клонировать репозиторий:
```bash
git clone https://github.com/ivasnev/FinFlow.git
cd FinFlow/FinFlowBackend/ff-id
```

2. Настроить конфигурацию:
```bash
cp config/config.yaml config/config.yaml.local
# Отредактировать config.yaml.local под свои настройки
```

Основные параметры конфигурации:
- `server.port` - порт сервиса (по умолчанию 8083)
- `postgres.*` - параметры подключения к PostgreSQL
- `redis.*` - параметры подключения к Redis
- `auth.*` - настройки интеграции с ff-auth
- `file_service.*` - настройки интеграции с ff-files
- `tvm.*` - настройки TVM для межсервисной авторизации

### Запуск с Docker Compose

```bash
docker-compose up -d
```

Это запустит:
- Сервис ff-id
- PostgreSQL (если не используется внешний)
- Redis (если не используется внешний)

### Локальный запуск

1. Убедитесь, что PostgreSQL и Redis запущены:
```bash
# PostgreSQL
psql -U postgres -c "CREATE DATABASE ff_id;"

# Redis (если не запущен)
redis-server
```

2. Примените миграции:
```bash
# Используя golang-migrate
migrate -path internal/repository/migrations -database "postgres://postgres:postgres@localhost:5432/ff_id?sslmode=disable" up
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

Сервис будет доступен по адресу `http://localhost:8083`

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

Бинарный файл будет создан в `bin/ff-id`.

### Запуск тестов

```bash
# Все тесты
make test

# С покрытием кода
make test-cov
```

### Генерация моков

```bash
make mock-gen
```

### Очистка сгенерированных файлов

```bash
make clean
```

## Конфигурация

Основные параметры конфигурации в `config/config.yaml`:

```yaml
server:
  port: 8083

postgres:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: ff_id

redis:
  host: localhost
  port: 6379
  password: ""

auth:
  host: http://localhost
  port: 8084
  update_interval: 10  # Интервал обновления публичных ключей (в секундах)

file_service:
  base_url: http://localhost:8082
  service_id: 2

tvm:
  base_url: http://localhost:8081
  service_id: 2
  service_secret: "..."  # Секрет для TVM

migrations:
  path: migrations

logger:
  level: info  # silent, error, warn, info
```

## Миграции базы данных

Миграции находятся в `internal/repository/migrations/`. Для применения миграций используйте `golang-migrate`:

```bash
# Применить все миграции
migrate -path internal/repository/migrations -database "postgres://user:pass@localhost:5432/ff_id?sslmode=disable" up

# Откатить последнюю миграцию
migrate -path internal/repository/migrations -database "postgres://user:pass@localhost:5432/ff_id?sslmode=disable" down 1
```

## Мониторинг и логирование

Сервис использует структурированное логирование с настраиваемыми уровнями:
- `silent` - без логов
- `error` - только ошибки
- `warn` - предупреждения и ошибки
- `info` - информационные сообщения, предупреждения и ошибки (по умолчанию)

## Интеграция с другими сервисами

### Регистрация пользователя

При регистрации пользователя в ff-auth, ff-auth вызывает внутренний endpoint ff-id для создания профиля:

```
POST /api/v1/users/register
X-TVM-Ticket: {ticket}
```

### Получение данных пользователей

Другие сервисы (например, ff-split) могут получать данные о пользователях:

```
GET /api/v1/users/ids?ids=1,2,3
X-TVM-Ticket: {ticket}
```

## Безопасность

- Все пользовательские endpoints защищены JWT токенами
- Внутренние endpoints защищены TVM тикетами
- Пароли не хранятся в этом сервисе (управляются в ff-auth)
- Данные пользователей кэшируются в Redis для повышения производительности
- CORS настроен для безопасного взаимодействия с фронтендом

## Лицензия

MIT
