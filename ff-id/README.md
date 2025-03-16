# FF-ID Service

Сервис управления пользователями для проекта FinFlow. Обеспечивает функционал регистрации, аутентификации и управления профилями пользователей.

## Функциональность

- Регистрация пользователей (email + пароль)
- Аутентификация с использованием JWT (access + refresh tokens)
- Управление профилем пользователя
- Ролевая система доступа (RBAC)
- Верификация email/телефона
- Восстановление пароля

## Технологии

- Go 1.21+
- PostgreSQL
- Redis
- Gin Web Framework
- GORM
- JWT

## Зависимости

- PostgreSQL 14+
- Redis 6+

## Установка и запуск

1. Клонируйте репозиторий:
```bash
git clone https://github.com/ivasnev/FinFlow.git
cd FinFlow/ff-id
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

### Публичные эндпоинты

- `POST /register` - Регистрация нового пользователя
- `POST /login` - Аутентификация пользователя
- `POST /refresh` - Обновление access token

### Защищенные эндпоинты (требуют JWT)

- `GET /profile` - Получение данных профиля
- `PUT /profile` - Обновление данных профиля
- `DELETE /profile` - Удаление профиля

### Административные эндпоинты

- Требуют роль `admin`
- Доступны по префиксу `/admin`

## Структура проекта

```
ff-id/
├── cmd/
│   └── app/
│       └── main.go
├── internal/
│   ├── config/
│   ├── handler/
│   ├── middleware/
│   ├── models/
│   ├── repository/
│   └── service/
├── go.mod
├── go.sum
└── README.md
```

## Безопасность

- Пароли хешируются с использованием bcrypt
- Используется JWT для аутентификации
- Поддерживается RBAC для авторизации
- Реализована защита от основных атак (CSRF, XSS) 