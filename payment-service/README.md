# payment-service

Сервис оплаты заказов. Хранит платежи в своей PostgreSQL-БД. Платёжный шлюз симулируется:
положительная сумма — платёж проводится успешно, нулевая/отрицательная — отклоняется.

## Эндпоинты

- `GET /health` — проверка живости
- `POST /api/payments` — провести оплату (`{"order_id": "...", "amount": 1999.90}`)
- `GET /api/payments/{id}` — статус платежа по id
- `GET /api/payments?order_id=...` — платежи по заказу (без параметра — все платежи)
- `GET /swagger/index.html` — Swagger UI

## Переменные окружения

| Переменная     | По умолчанию                                                       |
|----------------|------------------------------------------------------------------------|
| `PORT`         | `8083`                                                                  |
| `DATABASE_URL` | `postgres://payment:payment@localhost:5435/payment?sslmode=disable`    |

## Запуск локально

```bash
go run . migrate
go run .
```

## Запуск в Docker (вместе со всеми сервисами)

Из корня репозитория:

```bash
docker compose build payment-service
docker compose up -d --wait payment-db
docker compose run --rm payment-service migrate
docker compose up -d --wait payment-service
```

## Обновление Swagger-доков после правки аннотаций

```bash
go run github.com/swaggo/swag/cmd/swag init -g main.go -o docs
```
