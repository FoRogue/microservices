# microservices

Учебный репозиторий по курсу микросервисов.

## part1

Первая лаба: 3 независимых микросервиса на Go, каждый со своей БД и Docker-контейнером, без связи друг с другом. Swagger (OpenAPI) на каждом сервисе.

- `auth-service` — авторизация/регистрация пользователей
- `catalog-service` — каталог товаров + корзина
- `payment-service` — оплата

Подробности по каждому сервису (эндпоинты, переменные окружения) — в README внутри его директории.

### Запуск всех сервисов

```bash
docker compose up --build
```

После старта:

- auth-service: <http://localhost:8081/swagger/index.html>
- catalog-service: <http://localhost:8082/swagger/index.html>
- payment-service: <http://localhost:8083/swagger/index.html>

### Запуск одного сервиса локально (без Docker)

Нужен запущенный Postgres, см. `DATABASE_URL` по умолчанию в README сервиса.

```bash
cd auth-service && go run .
```
