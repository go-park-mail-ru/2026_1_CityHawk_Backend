# 2026_1_CityHawk_Backend
Репозиторий backend команды CityHawk

## Старт проекта

```bash
go run .
```

Сервер по умолчанию запускается на `:8080`.

## База данных

Документация по структуре БД лежит в:
- `db/normalized/relations.md`
- `db/normalized/er_diagram.md`
- `db/normalized/schema.sql`
- `db/normalized/seed.sql`

Runnable migrations лежат в:
- `db/migrations/0001_init.up.sql`
- `db/migrations/0001_init.down.sql`
- `db/migrations/0002_seed.up.sql`
- `db/migrations/0002_seed.down.sql`
- `db/migrations/0003_search_trgm.up.sql`
- `db/migrations/0003_search_trgm.down.sql`

Требования к окружению для локального запуска:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=cityhawk
DB_PASSWORD=cityhawk
DB_NAME=cityhawk
DB_SSLMODE=disable
```

Применить схему и сиды можно так:

```bash
make db-schema
make db-seed
```

Или напрямую через `psql`:

```bash
psql "postgres://cityhawk:cityhawk@localhost:5432/cityhawk?sslmode=disable" -f db/migrations/0001_init.up.sql
psql "postgres://cityhawk:cityhawk@localhost:5432/cityhawk?sslmode=disable" -f db/migrations/0003_search_trgm.up.sql
psql "postgres://cityhawk:cityhawk@localhost:5432/cityhawk?sslmode=disable" -f db/migrations/0002_seed.up.sql
```

Для полного сброса:

```bash
make db-reset
```

## API

### `GET /api/health`

Проверка, что сервис жив.

- Тело запроса: нет
- Успешный ответ `200`:

```text
ok
```

### `POST /api/auth/register`

Регистрация пользователя с установкой access/refresh cookie.

- Тело запроса:

```json
{
  "email": "user@example.com",
  "password": "secret123",
  "username": "cityhawk_user"
}
```

- Успешный ответ `201`:

```json
{
  "message": "registration successful"
}
```

- Возможные ошибки:
  - `400`: `{"error":"invalid json"}` / ошибки валидации полей (`email`, `password`, `username`)
  - `409`: `{"error":"email already exists"}`
  - `500`: `{"error":"failed to issue tokens"}` или `{"error":"internal error"}`

### `POST /api/auth/login`

Логин пользователя с установкой access/refresh cookie.

- Тело запроса:

```json
{
  "email": "user@example.com",
  "password": "secret123"
}
```

- Успешный ответ `200`:

```json
{
  "message": "login successful"
}
```

- Возможные ошибки:
  - `400`: `{"error":"invalid json"}` / ошибки валидации полей (`email`, `password`)
  - `401`: `{"error":"invalid credentials"}`
  - `500`: `{"error":"failed to issue tokens"}`

### `POST /api/auth/refresh`

Обновление access токена по refresh cookie (с ротацией refresh).

- Тело запроса: нет
- Требование: cookie `refresh_token` должна быть передана браузером.
- Успешный ответ `200`:

```json
{
  "access_token": "<jwt>"
}
```

- Возможные ошибки:
  - `401`: `{"error":"missing refresh token"}`
  - `401`: `{"error":"invalid refresh token"}`

### `POST /api/auth/logout`

Выход пользователя: отзыв refresh сессии и очистка refresh cookie.

- Тело запроса: нет
- Требование: cookie `refresh_token`.
- Успешный ответ `200`:

```json
{
  "message": "logout successful"
}
```

- Возможные ошибки:
  - `401`: `{"error":"missing refresh token"}`

### `GET /api/me`

Возвращает текущего пользователя.

- Тело запроса: нет
- Требование: cookie `access_token`.

```text
Cookie: access_token=<jwt>
```

- Успешный ответ `200`:

```json
{
  "id": "20260309013000.000000000",
  "email": "user@example.com"
}
```

- Возможные ошибки:
  - `401`: `{"error":"missing access token"}`
  - `401`: `{"error":"invalid access token"}`
  - `401`: `{"error":"unauthorized"}`
  - `401`: `{"error":"user not found"}`

### `GET /api/events`

Список всех карточек мероприятий.

- Тело запроса: нет
- Успешный ответ `200`:

```json
{
      "items": [
    {
      "id": "uuid",
      "title": "Rock concert",
      "shortDescription": "Best rock night",
      "coverImageUrl": "https://example.com/event.jpg",
      "tags": [],
      "nextSession": {
        "startAt": "2026-03-30T19:00:00Z",
        "place": {
          "name": "Arena",
          "addressLine": "Lenina 1"
        }
      }
    }
  ]
}
```

- Возможные ошибки:
  - `400`: `{"error":"Validation failed"}`
  - `405`: `{"error":"method not allowed"}`

### `GET /api/events/{id}`

Полная информация о выбранном мероприятии.

- Тело запроса: нет
- Пример: `GET /api/events/futurione`
- Успешный ответ `200`:

```json
{
  "id": "uuid",
  "title": "Rock concert",
  "shortDescription": "Best rock night",
  "fullDescription": "Long description",
  "ageLimit": 18,
  "sourceUrl": "https://example.com",
  "author": {
    "id": "uuid",
    "username": "Alice",
    "avatarUrl": null
  },
  "categories": [],
  "tags": [],
  "images": [],
  "sessions": [],
  "createdAt": "2026-03-20T10:00:00Z",
  "updatedAt": "2026-03-22T10:00:00Z",
  "isFavorite": false,
  "isOwner": true
}
```

- Возможные ошибки:
  - `404`: `{"error":"place not found"}`
  - `405`: `{"error":"method not allowed"}`

### `GET /places/category/{category}`

Отфильтрованные карточки по категориям (краткий формат).

- Тело запроса: нет
- Пример: `GET /places/category/park`
- Успешный ответ `200`:

```json
{
    "items": [
        {
            "id": "vdnh",
            "title": "ВДНХ",
            "categories": ["park"],
            "short_description": "Большой парковый и выставочный комплекс для прогулок и активностей.",
            "address": "Москва, проспект Мира, 119",
            "image_url": "images/vdnh.jpg"
        },
        {
            "id": "zaryadye",
            "title": "Парк Зарядье",
            "categories": ["park", "photo"],
            "short_description": "Центральный парк с панорамным мостом и видами на Кремль.",
            "address": "Москва, ул. Варварка, 6",
            "image_url": "images/zaryadye.jpg"
        }
    ]
}
```

- Возможные ошибки:
  - `404`: `{"error":"category not found"}`
  - `405`: `{"error":"method not allowed"}`

### `GET /places/best`

Топ карточек мест по `like_count` (по убыванию), максимум `8` элементов.

- Тело запроса: нет
- Успешный ответ `200`:

```json
{
  "items": [
    {
      "id": "zaryadye",
      "title": "Парк Зарядье",
      "categories": ["park", "photo"],
      "like_count": 150,
      "short_description": "Центральный парк с панорамным мостом и видами на Кремль.",
      "address": "Москва, ул. Варварка, 6",
      "image_url": "images/zaryadye.jpg"
    }
  ]
}
```

- Возможные ошибки:
  - `405`: `{"error":"method not allowed"}`

## JWT

- `access_token`: JWT (HS256), в `sub` кладется `user_id`, передается через `HttpOnly` cookie `access_token`.
- `refresh_token`: opaque-токен, хранится на сервере (in-memory) в хешированном виде и связан с `user_id`, передается через `HttpOnly` cookie `refresh_token`.
- `POST /auth/refresh` читает refresh из cookie, выполняет ротацию и возвращает новый `access_token` (новые access/refresh снова кладутся в cookie).
- `POST /auth/logout` читает refresh из cookie и отзывает сессию.

## Переменные окружения

- `JWT_SECRET`
- `ACCESS_TOKEN_TTL` (по умолчанию `15m`)
- `REFRESH_TOKEN_TTL` (по умолчанию `168h`)
