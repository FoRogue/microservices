# auth-service

Сервис авторизации/регистрации пользователей. Хранит пользователей в своей PostgreSQL-БД, выдаёт JWT-токены.

## Эндпоинты

- `GET /health` — проверка живости
- `POST /api/register` — регистрация (`{"username": "...", "password": "..."}`)
- `POST /api/login` — вход, возвращает `{"token": "..."}`
- `GET /api/me` — данные текущего пользователя по заголовку `Authorization: Bearer <token>`
- `GET /swagger/index.html` — Swagger UI

## Переменные окружения

| Переменная     | По умолчанию                                                  |
|----------------|----------------------------------------------------------------|
| `PORT`         | `8081`                                                          |
| `DATABASE_URL` | `postgres://auth:auth@localhost:5433/auth?sslmode=disable`      |
| `JWT_SECRET`   | `dev-secret-change-me`                                          |

## Запуск локально

```bash
go run . migrate
go run .
```

## Запуск в Docker (вместе со всеми сервисами)

Из корня репозитория:

```bash
docker compose build auth-service
docker compose up -d --wait auth-db
docker compose run --rm auth-service migrate
docker compose up -d --wait auth-service
```

## Обновление Swagger-доков после правки аннотаций

```bash
go run github.com/swaggo/swag/cmd/swag init -g main.go -o docs
```
