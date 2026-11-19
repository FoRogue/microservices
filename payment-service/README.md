# payment-service

Заказы отправляют запросы через `payments.requested`. Оплата публикует
`PaymentSucceeded` или `PaymentFailed` в `payments.completed`. REST API
доступен для просмотра платежей и ручной симуляции. Префикс ключей
`kafka-order:` используется обработчиком событий.

Сервис оплаты заказов. Хранит платежи в своей PostgreSQL-БД. Платёжный шлюз симулируется:
положительная сумма — платёж проводится успешно, нулевая/отрицательная — отклоняется.

## Эндпоинты

- `GET /health` — проверка живости
- `POST /api/payments` — провести оплату (`{"order_id": "...", "amount": 1999.90}`)
- `GET /api/payments/{id}` — статус платежа по id
- `GET /api/payments?order_id=...` — платежи по заказу (без параметра — все платежи)
- `GET /swagger/index.html` — Swagger UI

Для повторов `POST /api/payments` можно передать заголовок `Idempotency-Key`.
Повтор с тем же ключом и данными возвращает исходный платёж. Если ключ уже
привязан к другому заказу или сумме, сервис возвращает `409 Conflict`.

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
