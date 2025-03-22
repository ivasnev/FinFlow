# TVM API Documentation

## Общее описание

TVM (Ticket Vending Machine) - это сервис для управления доступом между сервисами с использованием тикетов.

## Аутентификация

Все запросы к API должны содержать заголовок `Content-Type: application/json`.

Для аутентификации между сервисами используется заголовок `X-FF-Service-Ticket` в формате:
```
X-FF-Service-Ticket: serv:{from_id_serv in base64}:{ticket in base64}
```

где:
- `from_id_serv` - ID сервиса-отправителя в base64
- `ticket` - тикет в base64

## Endpoints

### Создание сервиса

```http
POST /service
```

Создает новый сервис и возвращает его идентификатор и ключи.

#### Request Body

```json
{
    "name": "string" // Название сервиса
}
```

#### Response

```json
{
    "id": "number", // ID сервиса
    "name": "string", // Название сервиса
    "public_key": "string", // Публичный ключ в base64
    "private_key": "string" // Приватный ключ в base64 (важно сохранить!)
}
```

### Получение публичного ключа сервиса

```http
GET /service/{id}/key
```

Возвращает публичный ключ сервиса по его ID.

#### Response

```json
{
    "public_key": "string" // Публичный ключ в base64
}
```

### Предоставление доступа

```http
POST /access/grant
```

Предоставляет доступ от одного сервиса к другому.

#### Request Body

```json
{
    "from": "number", // ID сервиса-отправителя
    "to": "number" // ID сервиса-получателя
}
```

#### Response

```json
{
    "message": "access granted"
}
```

### Отзыв доступа

```http
POST /access/revoke
```

Отзывает доступ от одного сервиса к другому.

#### Request Body

```json
{
    "from": "number", // ID сервиса-отправителя
    "to": "number" // ID сервиса-получателя
}
```

#### Response

```json
{
    "message": "access revoked"
}
```

### Генерация тикета

```http
POST /ticket
```

Генерирует тикет для доступа от одного сервиса к другому.

#### Request Body

```json
{
    "from": "number", // ID сервиса-отправителя
    "to": "number", // ID сервиса-получателя
    "secret": "string" // Приватный ключ сервиса-отправителя в base64
}
```

#### Response

```json
{
    "from": "number", // ID сервиса-отправителя
    "to": "number", // ID сервиса-получателя
    "ttl": "number", // Время жизни тикета (Unix timestamp)
    "signature": "string", // Подпись тикета в base64
    "metadata": "string" // Метаданные в формате JSON
}
```

### Генерация тикета для разработки

```http
POST /dev/ticket
```

Генерирует тикет для разработки без необходимости знания секрета сервиса.

#### Request Body

```json
{
    "from": "number", // ID сервиса-отправителя
    "to": "number" // ID сервиса-получателя
}
```

#### Response

```json
{
    "from": "number", // ID сервиса-отправителя
    "to": "number", // ID сервиса-получателя
    "ttl": "number", // Время жизни тикета (Unix timestamp)
    "signature": "string", // Подпись тикета в base64
    "metadata": "string" // Метаданные в формате JSON
}
```

> **Примечание**: Этот эндпоинт предназначен только для разработки и тестирования. В продакшене он должен быть отключен или защищен дополнительной аутентификацией.

## Коды ошибок

- `400 Bad Request` - неверный формат запроса
- `401 Unauthorized` - неверный секрет
- `403 Forbidden` - доступ запрещен
- `404 Not Found` - сервис не найден
- `500 Internal Server Error` - внутренняя ошибка сервера

## Примеры использования

### Создание сервиса

```bash
curl -X POST http://localhost:8080/service \
  -H "Content-Type: application/json" \
  -d '{"name": "test-service"}'
```

### Генерация тикета

```bash
curl -X POST http://localhost:8080/ticket \
  -H "Content-Type: application/json" \
  -d '{
    "from": 1,
    "to": 2,
    "secret": "your-private-key"
  }'
``` 