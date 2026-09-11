# catalog-service

Сервис каталога товаров и корзины покупок. Хранит данные в своей PostgreSQL-БД.

## Эндпоинты

- `GET /health` — проверка живости
- `GET /api/products` — список товаров
- `POST /api/products` — создать товар (`{"name": "...", "description": "...", "price": 100.0}`)
- `GET /api/products/{id}` — получить товар
- `PUT /api/products/{id}` — обновить товар
- `DELETE /api/products/{id}` — удалить товар
- `GET /api/cart?user_id=...` — содержимое корзины пользователя
- `POST /api/cart/items` — добавить товар в корзину (`{"user_id": "...", "product_id": 1, "quantity": 2}`)
- `PUT /api/cart/items/{id}` — изменить количество (`{"quantity": 3}`)
- `DELETE /api/cart/items/{id}` — удалить позицию из корзины
- `GET /swagger/index.html` — Swagger UI

## Переменные окружения

| Переменная     | По умолчанию                                                        |
|----------------|-----------------------------------------------------------------------|
| `PORT`         | `8082`                                                                 |
| `DATABASE_URL` | `postgres://catalog:catalog@localhost:5434/catalog?sslmode=disable`   |

## Запуск локально

```bash
go run .
```

## Запуск в Docker (вместе со всеми сервисами)

Из корня репозитория:

```bash
docker compose up --build catalog-service catalog-db
```

## Обновление Swagger-доков после правки аннотаций

```bash
go run github.com/swaggo/swag/cmd/swag init -g main.go -o docs
```
