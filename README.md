# microservices

Учебный репозиторий по курсу микросервисов.

## part1

Первая лаба: 3 независимых микросервиса на Go, каждый со своей БД и Docker-контейнером, без связи друг с другом. Swagger (OpenAPI) на каждом сервисе.

- `auth-service` — авторизация/регистрация пользователей
- `catalog-service` — каталог товаров + корзина
- `payment-service` — оплата

Подробности по каждому сервису (эндпоинты, переменные окружения) — в README внутри его директории.

### Запуск всех сервисов

```powershell
docker compose build
docker compose up -d --wait auth-db catalog-db payment-db
docker compose run --rm auth-service migrate
docker compose run --rm catalog-service migrate
docker compose run --rm payment-service migrate
docker compose up -d --wait
```

После старта:

- auth-service: <http://localhost:8081/swagger/index.html>
- catalog-service: <http://localhost:8082/swagger/index.html>
- payment-service: <http://localhost:8083/swagger/index.html>

### Проверка через Compose

Из корня проекта в PowerShell выполните `./smoke-test.ps1`. Скрипт проверяет
регистрацию и вход, добавление товара в корзину, запись платежа и создание товара
при остановленном `payment-service`. После проверки он запускает платёжный сервис
снова и читает сохранённый платёж. Контейнеры и тестовые записи остаются в БД.

### Работа с базой

SQL-запросы находятся в `internal/repository.go` каждого сервиса. Бизнес-логика
авторизации и расчёта статуса платежа находится в `internal/service.go`.
Команда `migrate` применяет миграции из `internal/migrations/` к БД сервиса
и записывает применённые версии в `schema_migrations`. Выполняйте её перед
первым запуском и после обновлений, меняющих схему. Начальная миграция
поддерживает базы, созданные старой версией приложения. Для изменения схемы
добавляйте новый SQL-файл с очередным номером, например `002_add_index.sql`.

### Запуск одного сервиса локально (без Docker)

Нужен запущенный Postgres, см. `DATABASE_URL` по умолчанию в README сервиса.

```bash
cd auth-service
go run . migrate
go run .
```
