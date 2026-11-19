# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Статус репозитория

Учебный репозиторий по курсу микросервисов. В **part2** исходные четыре сервиса используют Kafka для оплаты.

- `auth-service` (порт 8081) — регистрация/логин, JWT, своя БД `auth`
- `catalog-service` (порт 8082) — товары + корзина, своя БД `catalog`
- `payment-service` (порт 8083) — оплата заказов (симулированный шлюз), своя БД `payment`
- `order-service` (порт 8084) — оформление и состояние заказов, своя БД `orders`

У каждого сервиса свой Go-модуль (`go.mod`), своя PostgreSQL-БД и Dockerfile.
Сервис заказов обращается к авторизации, каталогу и оплате по HTTP.
Склад получает события заказов и закупок из Kafka. Контракт событий и транспорт
находятся в локальном Go-модуле `eventing/`.

## Архитектура каждого сервиса

Сервисы построены по одному шаблону:

- `main.go` — точка входа: грузит конфиг из env, открывает БД, монтирует роуты и Swagger UI (`/swagger/index.html`).
- `internal/config.go` — конфиг из переменных окружения (`PORT`, `DATABASE_URL`, для auth ещё `JWT_SECRET`).
- `internal/db.go` — подключение к PostgreSQL через `database/sql` и `pgx`.
- `internal/repository.go` — SQL-запросы и чтение результатов.
- `internal/migrations/` — версионируемые SQL-миграции для команды `migrate`.
- `internal/handlers.go` (+ `products.go`/`cart.go` в catalog-service) — обработчики на стандартном `net/http` (Go 1.22+ `ServeMux` с методами и `{id}`-паттернами в путях), с swaggo-аннотациями (`@Summary`, `@Router` и т.д.) прямо над функциями.
- `internal/models.go` — DTO и доменные структуры, используются одновременно как JSON-модели и как источник схем для Swagger.
- `docs/` — генерируется командой `swag init`, **не редактировать руками**.
- `tools.go` (тег `//go:build tools`) — фиктивный импорт `swag/cmd/swag`, чтобы `go mod tidy` не вычищал его из `go.sum` (утилита запускается через `go run`, не импортируется в коде).

Сервис заказов проверяет токены через `auth-service` и ограничивает чтение и
оплату владельцем заказа. API каталога и платежей доступны напрямую для демо.

## Частые команды

Все команды выполняются из директории конкретного сервиса.

```bash
go build ./...              # собрать
go vet ./...                 # статический анализ
go run . migrate             # подготовить схему PostgreSQL
go run .                     # запустить локально
go run github.com/swaggo/swag/cmd/swag init -g main.go -o docs   # перегенерировать Swagger после правки аннотаций
```

Из корня репозитория в PowerShell:

```powershell
./start.ps1
./smoke-test.ps1
```

Запускайте модульные тесты командой `go test ./...` из каждого модуля.

## Работа с Swagger

Аннотации живут прямо над HTTP-хендлерами в `internal/*.go`. После любого изменения сигнатур запросов/ответов или добавления эндпоинта нужно перегенерировать `docs/` командой `swag init` (см. выше) — иначе `/swagger/index.html` покажет устаревшую спецификацию.
