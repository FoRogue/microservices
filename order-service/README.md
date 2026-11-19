# order-service

Сервис хранит заказы пользователей и их статусы в собственной PostgreSQL.
При создании он проверяет токен через auth-service и получает текущие данные
товаров через catalog-service. Название и цена каждой позиции сохраняются
в заказе. Запрос оплаты создаёт `PaymentRequested` с сохранённой суммой
в outbox. Фоновый издатель отправляет его в Kafka. Результат из
`payments.completed` обновляет заказ до `PAID` или `PAYMENT_FAILED`.

| Метод | Путь | Результат |
| --- | --- | --- |
| GET | `/health` | Проверка доступности процесса |
| POST | `/api/orders` | Новый заказ в `PENDING_PAYMENT` |
| GET | `/api/orders/{id}` | Заказ текущего пользователя |
| POST | `/api/orders/{id}/pay` | `202` для оплаты в очереди, `200` для завершённой |

Для маршрутов заказов передайте `Authorization: Bearer <token>`.
Заказ остаётся в `PENDING_PAYMENT` до обработки результата. Повторный запрос
использует ту же запись outbox. Нулевая сумма приводит к `PAYMENT_FAILED`.

Настройки для локального запуска:

| Переменная | Значение по умолчанию |
| --- | --- |
| `PORT` | `8084` |
| `DATABASE_URL` | `postgres://orders:orders@localhost:5436/orders?sslmode=disable` |
| `AUTH_URL` | `http://localhost:8081` |
| `CATALOG_URL` | `http://localhost:8082` |
| `KAFKA_BROKERS` | `localhost:9092` |

Запуск всей системы и пример покупки описаны в [корневом README](../README.md).
