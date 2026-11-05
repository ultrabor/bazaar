# ISM-25 — Inventory ledger: checkStock / increaseBalance / decreaseBalance

- Linear: https://linear.app/ultrabor/issue/ISM-25/inventory-ledger-checkstock-increasebalance-decreasebalance
- Milestone: M2 — Catalog, Purchasing & Inventory
- Priority: Urgent
- Due: 2026-11-08
- Branch: `slwpoyninza32/ism-25-inventory-ledger-checkstock-increasebalance-decreasebalance`

## Зачем

Общие building blocks для:

- ISM-24 — `receiveGoods`
- ISM-27 — `transferGoods`
- ISM-28 — `runStocktake`
- ISM-29 — `sell`
- ISM-31 — `processReturn`

Порядок разблокировки после ISM-25: ISM-24 → ISM-29 → ISM-27 / ISM-28 / ISM-31.

## Scope из тикета

- [ ] `recordMovement(location, product, qty, reason)` — append-only
- [ ] `increaseBalance(location, product, qty)`
- [ ] `decreaseBalance(location, product, qty)`
- [ ] `checkStock(location, product, qty)` — блокирует отрицательный остаток

Требования:

- Таблицы `inventory_movements` (append-only) и `inventory_balances`.
- Остаток обновляется в той же транзакции, что и запись движения.
- Отрицательный остаток запрещён (политика компании по умолчанию).

Архивные локации (перенесено из ISM-7):

- `increaseBalance` / `decreaseBalance` отклоняют операции по архивной локации (`ErrLocationArchived` → 409 или 422).
- `checkStock` и история движений по архивной локации доступны для чтения.

---

## Шаг 0. Подготовка

```bash
git checkout -b slwpoyninza32/ism-25-inventory-ledger-checkstock-increasebalance-decreasebalance
```

Перевести тикет в **In Progress**.

## Шаг 1. Решить, как хранить количество

У `products.unit` есть значения `kg`, `liter`, `meter`, поэтому количество бывает дробным.

- **Рекомендация:** `NUMERIC(14,3)` в БД. В Go хранить в тысячных долях как `int64` (1.5 кг = `1500`). Не нужна библиотека decimal, нет ошибок округления float.
- Альтернатива: `shopspring/decimal`. Точнее, но это лишняя зависимость.

Решить сразу: от этого зависят сигнатуры всех функций.

## Шаг 2. Миграция `migrations/00010_inventory.sql`

1. `inventory_balances`
   - `company_id`, `location_id`, `product_id`
   - `quantity NUMERIC(14,3) NOT NULL CHECK (quantity >= 0)`
   - `updated_at TIMESTAMPTZ NOT NULL DEFAULT now()`
   - `PRIMARY KEY (location_id, product_id)`
   - FK на `locations` и `Products`
2. `inventory_movements`
   - `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`
   - `company_id`, `location_id`, `product_id`
   - `qty NUMERIC(14,3) NOT NULL CHECK (qty <> 0)`: знак задаёт направление (+ приход, − расход)
   - `reason TEXT NOT NULL CHECK (reason IN ('purchase','sale','return','transfer_in','transfer_out','stocktake','write_off'))`
   - `ref_type TEXT`, `ref_id UUID`: ссылка на документ (покупка, продажа и т.д.), нужна для ISM-24 / ISM-29
   - `created_by UUID`, `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`
   - Индекс `(location_id, product_id, created_at)` для истории
3. Append-only: триггер `BEFORE UPDATE OR DELETE ON inventory_movements` → `RAISE EXCEPTION`.
4. Блок `-- +goose Down`.

Проверка: `up` → `down` → `up` проходит без ошибок.

## Шаг 3. Модуль `internal/modules/inventory/`

Структура как у `location`: `model.go`, `repository.go`, `service.go`, `module.go`.

`model.go`:

- Типы `Movement`, `Balance`.
- Тип `Reason` с константами.
- Ошибки:
  - `ErrInsufficientStock`
  - `ErrLocationArchived`
  - `ErrInvalidQuantity` (qty ≤ 0)

## Шаг 4. Сначала тесты (RED)

Integration-тесты против Postgres, по образцу `internal/modules/role/integration_test.go`. Table-driven.

- [ ] `increaseBalance`: новая строка создаётся; существующая суммируется; в журнале есть запись.
- [ ] `decreaseBalance`: остаток уменьшился.
- [ ] `decreaseBalance` при нехватке → `ErrInsufficientStock`; остаток и журнал не изменились.
- [ ] `checkStock`: хватает → `nil`; не хватает → `ErrInsufficientStock`; строки нет → считать 0.
- [ ] qty ≤ 0 → `ErrInvalidQuantity`.
- [ ] Архивная локация: increase / decrease отклоняются, остаток не меняется.
- [ ] Архивная локация: `checkStock` и история читаются.
- [ ] Чужой `company_id` не видит и не меняет остатки.
- [ ] Concurrency: остаток 5, 10 горутин параллельно делают decrease по 1. Ровно 5 успехов, остаток 0. Запуск с `-race`.
- [ ] `UPDATE` / `DELETE` в `inventory_movements` падает.

## Шаг 5. Repository

Все методы работают через `r.db.Conn(ctx)`, поэтому используют транзакцию из ctx автоматически.

1. **`lockLocation(ctx, companyId, locationId)`**

   ```sql
   SELECT archived FROM locations
   WHERE id = $1 AND company_id = $2
   FOR SHARE
   ```

   Важно: не использовать `location.GetLocationById`. Он фильтрует `archived = false` и вернёт `ErrLocationNotFound` вместо `ErrLocationArchived`. `FOR SHARE` не даёт параллельно заархивировать склад во время операции.

2. **`insertMovement`**: простой `INSERT`.

3. **`increase`**: upsert

   ```sql
   INSERT INTO inventory_balances (company_id, location_id, product_id, quantity)
   VALUES ($1, $2, $3, $4)
   ON CONFLICT (location_id, product_id)
   DO UPDATE SET quantity   = inventory_balances.quantity + EXCLUDED.quantity,
                 updated_at = now()
   ```

4. **`decrease`**: атомарно, без отдельного SELECT, поэтому без гонки

   ```sql
   UPDATE inventory_balances
   SET quantity = quantity - $3, updated_at = now()
   WHERE location_id = $1 AND product_id = $2 AND quantity >= $3
   ```

   `RowsAffected() == 0` → `ErrInsufficientStock`.

5. **`getBalance`**: `pgx.ErrNoRows` → 0.

6. **`listMovements(location, product, limit, offset)`**: с пагинацией.

## Шаг 6. Service: публичные building blocks

Каждый метод оборачивается в `db.WithinTx`. Вложенный вызов переиспользует внешнюю транзакцию, поэтому `receiveGoods` / `sell` смогут вызывать их внутри своей транзакции.

- **`IncreaseBalance(ctx, companyId, locationId, productId, qty, reason, ref)`**
  1. Проверить qty > 0.
  2. `lockLocation`: archived → `ErrLocationArchived`.
  3. `increase`.
  4. `recordMovement(+qty)`.
- **`DecreaseBalance(...)`**
  1. Проверить qty > 0.
  2. `lockLocation`: archived → `ErrLocationArchived`.
  3. `decrease`.
  4. `recordMovement(-qty)`.
- **`CheckStock(ctx, companyId, locationId, productId, qty) error`**: только чтение, архивные локации разрешены.
- **`recordMovement`**: приватный. Писать в журнал только через increase / decrease, чтобы журнал не разошёлся с остатком.

Ошибки оборачивать: `fmt.Errorf("decrease balance: %w", err)`.

## Шаг 7. GREEN, затем рефакторинг

```bash
go test -race -cover ./internal/modules/inventory/...
```

Цель: покрытие ≥ 80%.

## Шаг 8 (опционально). HTTP только для чтения

- `GET /locations/{id}/stock`
- `GET /locations/{id}/movements`
- Проверка прав (см. ISM-116 location-scoped access).
- Маппинг ошибок: `ErrLocationArchived` → 409; `ErrInsufficientStock` → 409 (или 422).

Чтобы не раздувать тикет, можно вынести в отдельную задачу.

## Шаг 9. Перед коммитом

- [ ] `gofmt` / `goimports`, `go vet`
- [ ] Code review
- [ ] Отметить чекбоксы в Linear
- [ ] Коммит: `feat(inventory): add append-only ledger with balance checks`
