# API Documentation

## Общая информация

Базовый URL: `http://localhost:8080`

Все запросы должны содержать валидный TVM-билет в заголовке `X-TVM-Ticket`.

## Endpoints

### Загрузка файла

```http
POST /upload
Content-Type: multipart/form-data
X-TVM-Ticket: <ticket>

file: <file>
```

#### Параметры запроса

| Параметр | Тип | Описание |
|----------|-----|----------|
| file | File | Файл для загрузки |

#### Ответ

```json
{
    "id": 1,
    "name": "example.pdf",
    "size": 1024,
    "mime_type": "application/pdf",
    "bucket": "ff-files",
    "object_key": "2024/03/19/example.pdf",
    "created_at": "2024-03-19T12:00:00Z",
    "updated_at": "2024-03-19T12:00:00Z"
}
```

### Получение файла

```http
GET /files/{file_id}
X-TVM-Ticket: <ticket>
```

#### Параметры пути

| Параметр | Тип | Описание |
|----------|-----|----------|
| file_id | integer | ID файла |

#### Ответ

```
Content-Type: <mime_type>
Content-Disposition: attachment; filename=<filename>

<file_content>
```

### Получение метаданных файла

```http
GET /files/{file_id}/metadata
X-TVM-Ticket: <ticket>
```

#### Параметры пути

| Параметр | Тип | Описание |
|----------|-----|----------|
| file_id | integer | ID файла |

#### Ответ

```json
{
    "id": 1,
    "name": "example.pdf",
    "size": 1024,
    "mime_type": "application/pdf",
    "bucket": "ff-files",
    "object_key": "2024/03/19/example.pdf",
    "created_at": "2024-03-19T12:00:00Z",
    "updated_at": "2024-03-19T12:00:00Z"
}
```

### Удаление файла

```http
DELETE /files/{file_id}
X-TVM-Ticket: <ticket>
```

#### Параметры пути

| Параметр | Тип | Описание |
|----------|-----|----------|
| file_id | integer | ID файла |

#### Ответ

```json
{
    "message": "File deleted successfully"
}
```

### Генерация временной ссылки

```http
POST /files/{file_id}/temporary-url
X-TVM-Ticket: <ticket>
```

#### Параметры пути

| Параметр | Тип | Описание |
|----------|-----|----------|
| file_id | integer | ID файла |

#### Ответ

```json
{
    "url": "https://minio.example.com/ff-files/2024/03/19/example.pdf?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=...",
    "expires_at": "2024-03-19T13:00:00Z"
}
```

## Коды ответов

| Код | Описание |
|-----|----------|
| 200 | Успешный запрос |
| 400 | Неверный запрос |
| 401 | Неавторизованный доступ |
| 403 | Доступ запрещен |
| 404 | Файл не найден |
| 500 | Внутренняя ошибка сервера | 