# FF-Files

Сервис для хранения, загрузки, получения и удаления файлов с поддержкой TVM авторизации.

## Требования

- Docker
- Docker Compose

## Запуск

1. Клонируйте репозиторий:
```bash
git clone https://github.com/ivasnev/FinFlow.git
cd FinFlow
```

2. Запустите сервис и его зависимости:
```bash
cd ff-files
docker-compose up -d
```

Сервис будет доступен по адресу: http://localhost:8080

## API Endpoints

### Загрузка файла
- Метод: POST /upload
- Content-Type: multipart/form-data
- Параметры:
  - file: файл для загрузки
  - metadata: JSON с метаданными (опционально)
- Ответ: 201 Created + {"file_id": "uuid"}

### Получение файла
- Метод: GET /files/{file_id}
- Ответ: 200 OK + бинарные данные файла

### Удаление файла
- Метод: DELETE /files/{file_id}
- Ответ: 204 No Content

### Получение метаданных файла
- Метод: GET /files/{file_id}/metadata
- Ответ: 200 OK + {"size": 12345, "uploaded_at": "...", "owner": "user_id"}

### Генерация временной ссылки
- Метод: POST /files/{file_id}/temporary-url
- Параметры:
  - expires_in: время жизни ссылки в секундах
- Ответ: 200 OK + {"url": "https://..."}

## Авторизация

Все запросы должны содержать TVM-токен в заголовке `X-FF-Service-Ticket` и ID сервиса в заголовке `X-FF-Id-Service`.

## Разработка

1. Установите зависимости:
```bash
go mod download
```

2. Запустите тесты:
```bash
go test ./...
```

3. Соберите приложение:
```bash
go build -o server ./cmd/server
``` 