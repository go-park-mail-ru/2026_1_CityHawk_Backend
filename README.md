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

- `access_token`: короткоживущий, stateless, используется для доступа к API.
- `refresh_token`: долгоживущий, хранится на сервере (в памяти) в хешированном виде, используется для обновления `access_token`.
- При `POST /auth/refresh` выполняется ротация refresh-токена: старый токен отзывается, выдается новая пара токенов.
- `POST /auth/logout` отзывает переданный `refresh_token`.

## Переменные окружения

- `JWT_SECRET` (обязательно для production)
- `ACCESS_TOKEN_TTL` (по умолчанию `15m`)
- `REFRESH_TOKEN_TTL` (по умолчанию `168h`)
