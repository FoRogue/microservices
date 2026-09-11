# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Статус репозитория

Учебный репозиторий по курсу микросервисов. Реализован **part1**: три независимых микросервиса на Go.

- `auth-service` (порт 8081) — регистрация/логин, JWT, своя БД `auth`
- `catalog-service` (порт 8082) — товары + корзина, своя БД `catalog`
- `payment-service` (порт 8083) — оплата заказов (симулированный шлюз), своя БД `payment`

Сервисы **полностью независимы**: у каждого свой Go-модуль (`go.mod`), своя PostgreSQL-БД, свой Dockerfile, они не вызывают друг друга по сети.

## Архитектура каждого сервиса

Все три сервиса построены по одному шаблону:

- `main.go` — точка входа: грузит конфиг из env, открывает БД, монтирует роуты и Swagger UI (`/swagger/index.html`).
- `internal/config.go` — конфиг из переменных окружения (`PORT`, `DATABASE_URL`, для auth ещё `JWT_SECRET`).
- `internal/db.go` — открытие `*sql.DB` через `pgx` (драйвер `github.com/jackc/pgx/v5/stdlib`), ретраи пинга при старте, автоприменение схемы (`CREATE TABLE IF NOT EXISTS`) — отдельных миграций нет.
- `internal/handlers.go` (+ `products.go`/`cart.go` в catalog-service) — обработчики на стандартном `net/http` (Go 1.22+ `ServeMux` с методами и `{id}`-паттернами в путях), с swaggo-аннотациями (`@Summary`, `@Router` и т.д.) прямо над функциями.
- `internal/models.go` — DTO и доменные структуры, используются одновременно как JSON-модели и как источник схем для Swagger.
- `docs/` — генерируется командой `swag init`, **не редактировать руками**.
- `tools.go` (тег `//go:build tools`) — фиктивный импорт `swag/cmd/swag`, чтобы `go mod tidy` не вычищал его из `go.sum` (утилита запускается через `go run`, не импортируется в коде).

Аутентификация между сервисами не реализована — `catalog-service` и `payment-service` не проверяют JWT из `auth-service` (это независимые лабы по условию part1).

## Частые команды

Все команды — из директории конкретного сервиса (`auth-service/`, `catalog-service/` или `payment-service/`).

```bash
go build ./...              # собрать
go vet ./...                 # статический анализ
go run .                     # запустить локально (нужен Postgres, см. README сервиса)
go run github.com/swaggo/swag/cmd/swag init -g main.go -o docs   # перегенерировать Swagger после правки аннотаций
```

Из корня репозитория:

```bash
docker compose up --build           # поднять все 6 контейнеров (3 сервиса + 3 БД)
docker compose up --build auth-service auth-db   # поднять только один сервис с его БД
```

Тестов пока нет.

## Работа с Swagger

Аннотации живут прямо над HTTP-хендлерами в `internal/*.go`. После любого изменения сигнатур запросов/ответов или добавления эндпоинта нужно перегенерировать `docs/` командой `swag init` (см. выше) — иначе `/swagger/index.html` покажет устаревшую спецификацию.
