# 2026_1_CityHawk_Backend
Репозиторий backend команды CityHawk

## Старт проекта

```bash
go run .
```

Сервер по умолчанию запускается на `:8080`.

## API

### `GET /health`

Проверка, что сервис жив.

- Тело запроса: нет
- Успешный ответ `200`:

```text
ok
```

### `POST /auth/register`

Регистрация пользователя и выдача пары токенов.

- Тело запроса:

```json
{
  "email": "user@example.com",
  "password": "secret123"
}
```

- Успешный ответ `201`:

```json
{
  "access_token": "<jwt>",
  "refresh_token": "<opaque_token>",
  "token_type": "Bearer",
  "expires_in": 900
}
```

- Возможные ошибки:
  - `400`: `{"error":"invalid json"}` / `{"error":"email and password are required"}`
  - `409`: `{"error":"email already exists"}`
  - `500`: `{"error":"failed to issue tokens"}` или `{"error":"internal error"}`

### `POST /auth/login`

Логин пользователя и выдача пары токенов.

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
  "access_token": "<jwt>",
  "refresh_token": "<opaque_token>",
  "token_type": "Bearer",
  "expires_in": 900
}
```

- Возможные ошибки:
  - `400`: `{"error":"invalid json"}` / `{"error":"email and password are required"}`
  - `401`: `{"error":"invalid credentials"}`
  - `500`: `{"error":"failed to issue tokens"}`

### `POST /auth/refresh`

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

### `POST /auth/logout`

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

### `GET /me`

Возвращает текущего пользователя.

- Тело запроса: нет
- Заголовок:

```text
Authorization: Bearer <access_token>
```

- Успешный ответ `200`:

```json
{
  "id": "20260309013000.000000000",
  "email": "user@example.com"
}
```

- Возможные ошибки:
  - `401`: `{"error":"missing bearer token"}`
  - `401`: `{"error":"invalid access token"}`
  - `401`: `{"error":"unauthorized"}`
  - `401`: `{"error":"user not found"}`

### `GET /places`

Список всех карточек мест (краткий формат).

- Тело запроса: нет
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
      "image_url": "https://example.com/images/vdnh.jpg"
    }
  ]
}
```

- Возможные ошибки:
  - `405`: `{"error":"method not allowed"}`

### `GET /places/{id}`

Полная информация о выбранной карточке места.

- Тело запроса: нет
- Пример: `GET /places/vdnh`
- Успешный ответ `200`:

```json
{
    "item": {
        "id": "vdnh",
        "title": "ВДНХ",
        "categories": ["park"],
        "short_description": "Большой парковый и выставочный комплекс для прогулок и активностей.",
        "full_description": "ВДНХ объединяет павильоны, парковые зоны и культурные площадки. Подходит для долгих прогулок, семейных выходных и посещения временных выставок.",
        "address": "Москва, проспект Мира, 119",
        "image_url": "images/vdnh.jpg",
        "working_hours": "ежедневно, 10:00-22:00",
        "price_level": "free"
    }
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

- `access_token`: JWT (HS256), в `sub` кладется `user_id`.
- `refresh_token`: opaque-токен, хранится на сервере (in-memory) в хешированном виде и связан с `user_id`, возвращается в `register/login` и дополнительно ставится в `HttpOnly` cookie.
- `POST /auth/refresh` читает refresh из cookie, выполняет ротацию и возвращает новый `access_token` (новый refresh снова кладется в cookie).
- `POST /auth/logout` читает refresh из cookie и отзывает сессию.

## Переменные окружения

- `JWT_SECRET`
- `ACCESS_TOKEN_TTL` (по умолчанию `15m`)
- `REFRESH_TOKEN_TTL` (по умолчанию `168h`)
