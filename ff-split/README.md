# FinFlow Split Service (ff-split)

## Описание
FinFlow Split Service (ff-split) - это сервис для управления групповыми расходами и мероприятиями в экосистеме FinFlow. Сервис предоставляет API для создания и управления мероприятиями, учета расходов между участниками, и расчета задолженностей.

## Функциональность

- **Управление категориями мероприятий**
- **Создание и управление мероприятиями**
- **Отслеживание активностей в мероприятиях**
- **Управление транзакциями и расходами**
- **Расчет задолженностей между участниками**
- **Управление типами транзакций и иконками**

## Технический стек

- **Язык:** Go
- **Фреймворк:** Gin (HTTP-сервер)
- **База данных:** PostgreSQL
- **ORM:** GORM
- **Миграции:** golang-migrate
- **Развертывание:** Docker, Docker Compose

## Структура проекта

```
ff-split/
├── cmd/                # Точки входа в приложение
│   └── app/            # Основное приложение
├── config/             # Конфигурационные файлы
├── internal/           # Внутренний код приложения
│   ├── api/            # API компоненты
│   │   ├── handler/    # Обработчики HTTP-запросов
│   │   └── middleware/ # Промежуточные обработчики
│   ├── app/            # Ядро приложения
│   ├── common/         # Общие компоненты
│   │   └── config/     # Конфигурация приложения
│   ├── models/         # Модели данных
│   ├── repository/     # Слой доступа к данным
│   │   └── postgres/   # Реализация на PostgreSQL
│   └── service/        # Бизнес-логика
├── migrations/         # Миграции базы данных
└── pkg/                # Публичные пакеты
```

## API Endpoints

### Категории

```
GET  /api/v1/category        # Получить все категории
GET  /api/v1/category/{id}   # Получить категорию по ID
POST /api/v1/category        # Создать новую категорию
PUT  /api/v1/category/{id}   # Обновить категорию
DEL  /api/v1/category/{id}   # Удалить категорию
```

### Мероприятия

```
GET  /api/v1/events                     # Получить все мероприятия
GET  /api/v1/event/{id}                 # Получить мероприятие по ID
POST /api/v1/event                      # Создать новое мероприятие
PUT  /api/v1/event/{id}                 # Обновить мероприятие
DEL  /api/v1/event/{id}                 # Удалить мероприятие
```

### Активности мероприятий

```
GET  /api/v1/event/{id_event}/activity          # Получить все активности мероприятия
GET  /api/v1/event/{id_event}/activity/{id}     # Получить активность по ID
POST /api/v1/event/{id_event}/activity          # Создать новую активность
PUT  /api/v1/event/{id_event}/activity/{id}     # Обновить активность
DEL  /api/v1/event/{id_event}/activity/{id}     # Удалить активность
```

### Транзакции

```
GET  /api/v1/event/{id_event}/transactions      # Получить все транзакции мероприятия
GET  /api/v1/event/{id_event}/transactions/temporal  # Получить итоговые задолженности
```

### Управление (требуется роль service_admin)

```
GET  /api/v1/manage/transaction_type        # Получить все типы транзакций
POST /api/v1/manage/transaction_type        # Создать новый тип транзакции

GET  /api/v1/manage/icons                   # Получить все иконки
POST /api/v1/manage/icons                   # Загрузить новую иконку
```

## Запуск сервиса

### Предварительные требования

- Go 1.19 или выше
- PostgreSQL 13 или выше
- Docker и Docker Compose (опционально)

### Настройка окружения

1. Клонировать репозиторий
```
git clone https://github.com/ivasnev/FinFlow.git
cd FinFlow/ff-split
```

2. Настроить переменные окружения или файл конфигурации
```
cp config/config.example.yaml config/config.yaml
# Отредактировать config.yaml под свои настройки
```

### Запуск с Docker

```
docker-compose up -d
```

### Локальный запуск

1. Установить зависимости
```
go mod download
```

2. Применить миграции
```
make migrate-up
```

3. Запустить сервис
```
make run
```

## Разработка

### Сборка проекта
```
make build
```

### Запуск тестов
```
make test
```

### Запуск линтера
```
make lint
```

### Запуск в режиме разработки (hot-reload)
```
make dev
``` 