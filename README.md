# 2026_1_CityHawk_Backend
Репозиторий backend команды CityHawk

## Старт проекта

```bash
go run .
```

Сервер по умолчанию запускается на `:8080`.

## Тесты

В проекте используются:
- `gomock` для unit-тестов usecase и gateway-слоя
- `sqlmock` для SQL-ориентированных тестов

Важно: для стабильного расчета покрытия в этом проекте используется `Go 1.25.0` через `GOTOOLCHAIN=go1.25.0`, потому что на локальном `go1.26.0` `coverprofile` работает нестабильно. При первом запуске Go может автоматически скачать toolchain `1.25.0`.

Быстрый запуск всех тестов:

```bash
make test
```

Посчитать покрытие:

```bash
make coverage
```

Проверить, что покрытие не ниже 60%:

```bash
make coverage-check
```

Что делает `coverage`:
- гоняет `go test ./...`
- считает общее покрытие по проекту через `-coverpkg`
- не учитывает entrypoint-пакеты `cmd/...`
- исключает из итогового профиля сгенерированные файлы: `*_easyjson.go`, `*_mock.go`, `internal/mocks`, `pkg/pb`

После финального прогона в проекте покрытие составляет `60.1%`.

## Генерация кода

Для JSON-сериализации и десериализации HTTP DTO используется `easyjson`.

Перегенерировать easyjson-файлы:

```bash
make generate
```

Команда запускает `go run github.com/mailru/easyjson/easyjson` с нужными флагами для всех request/response DTO. Для request-структур включен `-disallow_unknown_fields`, чтобы строгая проверка неизвестных JSON-полей сохранялась после генерации.

## Микросервисы и gRPC

Проект можно запускать в двух режимах:

- `./cmd` - существующий HTTP backend, совместимый с текущим API;
- `./cmd/*-service` - отдельные gRPC-сервисы, собранные по proto-контрактам из `proto/cityhawk/*/v1`.

Доступные gRPC-сервисы:

```text
auth-service     :50051  cityhawk.auth.v1.AuthService
profile-service  :50052  cityhawk.profile.v1.ProfileService
events-service   :50053  cityhawk.events.v1.EventsService
support-service  :50054  cityhawk.support.v1.SupportService
social-service   :50055  cityhawk.social.v1.SocialService
```

Сгенерированные клиентские и серверные интерфейсы лежат в `pkg/pb/*/v1`, а серверные адаптеры - в `internal/*/delivery/grpc`.

Межсервисное общение по gRPC уже используется внутри системы: `support-service` ходит в `profile-service`
через `cityhawk.profile.v1.ProfileService/GetUser`, чтобы проверять автора тикета и его роль перед операциями
поддержки. В Docker Compose для этого `SUPPORT` контейнер получает `PROFILE_GRPC_ADDR=cityhawk-profile-service:50052`.

Локальная сборка всех бинарников:

```bash
make build-services
```

Запуск одного сервиса:

```bash
AUTH_GRPC_ADDR=:50051 go run ./cmd/auth-service
PROFILE_GRPC_ADDR=:50052 go run ./cmd/profile-service
EVENTS_GRPC_ADDR=:50053 go run ./cmd/events-service
SUPPORT_GRPC_ADDR=:50054 go run ./cmd/support-service
SOCIAL_GRPC_ADDR=:50055 go run ./cmd/social-service
```

Через Docker Compose:

```bash
docker compose up -d postgres photon cityhawk-auth-service cityhawk-profile-service cityhawk-events-service cityhawk-support-service cityhawk-social-service
```

## Мониторинг

Prometheus, Grafana, node-exporter и cAdvisor вынесены в отдельный Docker Compose profile, чтобы обычный запуск приложения не зависел от мониторинга:

```bash
docker compose --profile monitoring up -d
```

Grafana доступна на `http://localhost:3000` (`admin` / `admin` по умолчанию), Prometheus - на `http://localhost:9090`.

Dashboard `CityHawk RK Monitoring` автоматически подключается через provisioning и содержит:
- хиты, ошибки и тайминги запросов по всем HTTP/gRPC сервисам, методам и URL/RPC route;
- метрики всех микросервисов: `cityhawk-backend`, `cityhawk-auth-service`, `cityhawk-profile-service`, `cityhawk-events-service`, `cityhawk-support-service`, `cityhawk-social-service`;
- CPU, память и диск машины через `node-exporter`;
- CPU и память контейнеров через `cAdvisor`.

HTTP backend отдает `/metrics` на `:8080`. gRPC-сервисы отдают `/metrics` на отдельных HTTP-портах:

```text
auth-service     :9101
profile-service  :9102
events-service   :9103
support-service  :9104
social-service   :9105
```

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
- `db/migrations/0005_media_paths.up.sql`
- `db/migrations/0005_media_paths.down.sql`

Требования к окружению для локального запуска:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=cityhawk
DB_PASSWORD=cityhawk
DB_NAME=cityhawk
DB_SSLMODE=disable
PHOTON_ENABLED=false
PHOTON_BASE_URL=http://localhost:2322
PHOTON_REQUEST_TIMEOUT=5s
PHOTON_DEFAULT_COUNTRY=Russia
PHOTON_DEFAULT_TIMEZONE=Europe/Moscow
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
psql "postgres://cityhawk:cityhawk@localhost:5432/cityhawk?sslmode=disable" -f db/migrations/0005_media_paths.up.sql
psql "postgres://cityhawk:cityhawk@localhost:5432/cityhawk?sslmode=disable" -f db/migrations/0002_seed.up.sql
```

После `make db-seed` база будет заполнена 100 сидовыми карточками событий, и у каждой карточки будут связанные категории, теги, картинка и сессия. Дополнительно создаются сидовые коллекции и избранные события для тестирования API.

Для полного сброса:

```bash
make db-reset
```

## Photon Self-Hosted

Для адресных подсказок можно поднять локальный Photon через `docker compose`.

По умолчанию в `docker-compose.yml` уже добавлен сервис `photon`:

- порт: `2322`
- регион индекса: `russia`
- данные хранятся в docker volume `photon_data`
- backend внутри compose ходит в Photon по адресу `http://photon:2322`

Минимальные переменные окружения:

```env
PHOTON_ENABLED=true
PHOTON_BASE_URL=http://localhost:2322
PHOTON_REQUEST_TIMEOUT=5s
PHOTON_DEFAULT_COUNTRY=Russia
PHOTON_DEFAULT_TIMEZONE=Europe/Moscow
PHOTON_REGION=russia
PHOTON_UPDATE_STRATEGY=DISABLED
PHOTON_LOG_LEVEL=INFO
```

Запуск:

```bash
docker compose up -d photon postgres cityhawk-backend
```

Проверка Photon:

```bash
curl "http://localhost:2322/status"
curl "http://localhost:2322/api?q=ВДНХ&limit=5"
```

Проверка backend-подсказок:

```bash
curl "http://localhost:8080/api/place-suggestions?query=ВДНХ&limit=5"
```

Выбор подсказки и создание записи в таблице `place`:

1. Взять `token` из ответа `/api/place-suggestions`.
2. Отправить его в backend:

```bash
curl -X POST "http://localhost:8080/api/places/resolve" \
  -H "Content-Type: application/json" \
  -d '{"token":"<token-from-suggestion>"}'
```

Ответ вернет готовый `placeId`, который потом надо использовать в `sessions[].placeId` при создании события через `POST /api/events`.

Как должен работать frontend:

1. При вводе в поле `место` вызывать `GET /api/place-suggestions?query=<text>&limit=5`.
2. Показывать только выпадающий список подсказок.
3. Не разрешать сабмит формы, пока пользователь не выбрал один из вариантов.
4. После выбора сохранить у себя `token`.
5. Перед созданием события вызвать `POST /api/places/resolve`.
6. Взять из ответа `id` и отправить его как `placeId` в `sessions`.

Если пользователь после выбора снова меняет текст руками, выбранный `token` нужно сбросить и потребовать новое явное выбор из подсказок.

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

### `PATCH /api/me`

Частично обновляет профиль текущего пользователя. Для загрузки аватарки используйте `multipart/form-data` и поле файла `avatar`.

Пример:

```text
PATCH /api/me
Cookie: access_token=<jwt>
Content-Type: multipart/form-data

username=Alice
avatar=<binary file>
```

- Успешный ответ `200` возвращает обновленный профиль, а `avatarUrl` указывает на сохраненный файл вида `/uploads/avatars/...`.
- Возможные ошибки:
  - `400`: `{"error":"Validation failed"}`
  - `401`: `{"error":"Unauthorized"}`

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

### `POST /api/events`

Создает событие. Для загрузки картинок можно использовать `multipart/form-data` и поле файлов `images`.
Поля `categoryIds`, `tagIds`, `imageUrls`, `sessions` в этом случае передаются JSON-строками.

```text
POST /api/events
Cookie: access_token=<jwt>
Content-Type: multipart/form-data

title=New Event
shortDescription=Short text
fullDescription=Long event description
categoryIds=["music"]
# sessions опционален
# sessions=[{"placeId":"place-1","startAt":"2026-04-20T19:00:00Z","endAt":"2026-04-20T21:00:00Z","price":1200}]
images=<binary file>
images=<binary file>
```

- Успешный ответ `201`:
```json
{"id":"uuid"}
```
- Возможные ошибки:
  - `400`: `{"error":"Validation failed"}`
  - `401`: `{"error":"Unauthorized"}`

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
