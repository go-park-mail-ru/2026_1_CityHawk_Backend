# 2026_1_CityHawk_Backend
Репозиторий backend команды CityHawk

## Старт проекта

```bash
go run .
```

Сервер по умолчанию запускается на `:8080`.

## Ручки авторизации

- `POST /auth/register`
- `POST /auth/login`
- `POST /auth/refresh`
- `POST /auth/logout`
- `GET /me` (требует `Authorization: Bearer <access_token>`)
- `GET /health`

## JWT

- `access_token`: JWT (HS256), в `sub` кладется `user_id`.
- `refresh_token`: opaque-токен, хранится на сервере (in-memory) в хешированном виде и связан с `user_id`, возвращается в `register/login` и дополнительно ставится в `HttpOnly` cookie.
- `POST /auth/refresh` читает refresh из cookie, выполняет ротацию и возвращает новый `access_token` (новый refresh снова кладется в cookie).
- `POST /auth/logout` читает refresh из cookie и отзывает сессию.

## Переменные окружения

- `JWT_SECRET` 
- `ACCESS_TOKEN_TTL` (по умолчанию `15m`)
- `REFRESH_TOKEN_TTL` (по умолчанию `168h`)
