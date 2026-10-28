# Архитектура

Bazaar — модульный монолит. Один модуль — один Go-пакет в
`internal/modules/<name>`,
а слои — это файлы внутри пакета: `handler → service → repository`.

## Структура

```
cmd/api/main.go           конфиг, БД, миграции, HTTP-сервер
internal/
  app/                    сборка модулей и роутер
  modules/
    <module>/             auth, user, role, permission, product, category, location
      model.go            сущности и ошибки модуля
      repository.go       SQL (pgx), пишет только в таблицы модуля
      service.go          бизнес-логика и интерфейсы зависимостей
      handler.go          HTTP-хендлеры и тела запросов (*Body)
      routes.go           интерфейс для хендлеров, Handler, Routes
      module.go           New, Service(), Routes — всё, что нужно app
      *_test.go
  platform/               инфраструктура без бизнес-логики
    database/             пул pgx, TxManager, Querier
    httpx/                WriteError, WriteJSON, CurrentUser, swagger
    httpx/middleware/     Auth, Permission, Logger, Recover
    httpx/httpxtest/      хелперы для тестов хендлеров
    identity/             текущий пользователь запроса
    health/               /healthz, /readyz
    config/ logger/ worker/ support/
```

## Правила

1. **Слои внутри модуля.** Хендлер вызывает только `Service`, сервис —
   только `repository` и другие модули. Это файлы одного пакета, поэтому
   правило держится на код-ревью.
2. **Граница модуля — пакет.** Другие модули видят только экспортированное:
   `role.Service`, `role.ErrRoleNotAssignable`. Неэкспортированные функции и
   поля скрыты самим Go.
3. **Интерфейсы объявляет потребитель.** `user` описывает `roleService` из
   трёх нужных ему методов; `*role.Service` подходит под него сам.
   Конструкторы возвращают структуры.
4. **Направление зависимостей** закреплено в `.golangci.yml` (depguard):

   ```
   auth ───────► role
   user ───────► role
   permission ─► user
   category, location, product — независимы
   platform — не импортирует ни один модуль
   ```

   Новую связь сначала разрешают в `.golangci.yml`. Обратные связи
   (`role → user`) запрещены: Go не соберёт цикл импорта.
5. **Ошибки.** У модуля свои ошибки в `model.go`. Общие `ErrInvalidCred` и
   `ErrForbidden` лежат в `platform/support/apperror`.
6. **Транзакции.** `database.TxManager.WithinTx` кладёт транзакцию в ctx.
   Вызовы репозиториев и других модулей с этим ctx попадают в ту же
   транзакцию. Так операция из нескольких модулей (продажа: остатки,
   оплата, долг, аудит) остаётся атомарной.
7. **Таблицы.** Модуль пишет только в свои таблицы. Исключение — будущий
   `reporting`: он только читает таблицы других модулей.

## Как добавить модуль

1. Создать `internal/modules/<name>/` с файлами `model.go`, `repository.go`,
   `service.go`, `handler.go`, `routes.go`, `module.go`.
2. В `internal/app/app.go` вызвать `<name>.New(...)` и передать модуль в
   `NewRouter`.
3. Добавить правило модуля в `.golangci.yml`.
4. Перегенерировать swagger: `make swag_init`.

Если модуль разрастается (например, `inventory`), его делят на файлы по
темам: `ledger.go`, `fifo.go`, `transfer.go`.
