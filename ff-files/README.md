# FF-Files Service

FF-Files - это микросервис для управления файлами в экосистеме FinFlow. Сервис предоставляет возможности загрузки, хранения, получения и управления файлами с поддержкой метаданных и временных URL.

## Функциональность

### Загрузка файлов
- Загрузка файлов через форму
- Загрузка файлов по URL
- Валидация размера файла и MIME-типов
- Сохранение метаданных файла

### Получение файлов
- Получение файла по ID
- Получение метаданных файла
- Получение файла по временному URL

### Управление файлами
- Мягкое удаление файлов
- Автоматическая очистка удаленных файлов
- Управление временными URL

## API Endpoints

### Загрузка файлов
```
POST /upload
Content-Type: multipart/form-data
Body: file=@file.jpg

POST /upload/url
Content-Type: application/json
{
    "url": "https://example.com/file.jpg"
}
```

### Работа с файлами
```
GET /file/:id
GET /file/:id/meta
DELETE /file/:id
```

### Временные URL
```
POST /file/:id/url
Content-Type: application/json
{
    "duration": "24h"
}

GET /temp/:id
```

## Конфигурация

Сервис настраивается через конфигурационный файл со следующими параметрами:

```yaml
server:
  port: ":8082"

database:
  host: "localhost"
  port: "5432"
  user: "postgres"
  password: "postgres"
  dbname: "ff_files"

storage:
  basePath: "./storage"
  maxFileSize: 10485760  # 10MB
  allowedMimeTypes:
    - "image/jpeg"
    - "image/png"
    - "image/gif"
    - "application/pdf"
    - "text/plain"
  tempURLExpiration: "24h"
  softDeleteTimeout: "720h"  # 30 days
```

## Установка и запуск

1. Клонируйте репозиторий:
```bash
git clone https://github.com/ivasnev/FinFlow.git
cd FinFlow/ff-files
```

2. Установите зависимости:
```bash
go mod download
```

3. Создайте и настройте конфигурационный файл:
```bash
cp config.example.yaml config.yaml
```

4. Запустите сервис:
```bash
go run cmd/app/main.go
```

## Зависимости

- Go 1.21+
- PostgreSQL
- Gin Web Framework
- GORM
- Viper

## Безопасность

- Валидация размера файлов и MIME-типов
- Генерация уникальных идентификаторов для файлов
- Поддержка временных URL с ограниченным сроком действия
- Мягкое удаление файлов с отложенной очисткой

## Интеграция

Для интеграции с сервисом используйте предоставленные API endpoints. Пример интеграции:

```go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
)

func uploadFile(filename string, content []byte) error {
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    part, err := writer.CreateFormFile("file", filename)
    if err != nil {
        return err
    }
    _, err = part.Write(content)
    if err != nil {
        return err
    }
    writer.Close()

    req, err := http.NewRequest("POST", "http://localhost:8082/upload", body)
    if err != nil {
        return err
    }
    req.Header.Set("Content-Type", writer.FormDataContentType())

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    return nil
}
```

## Мониторинг

Сервис предоставляет базовое логирование операций и ошибок. В будущих версиях планируется добавление метрик и трейсинга. 