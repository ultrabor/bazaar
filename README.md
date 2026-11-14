# Bazaar

Backend for a small-business management system: wholesale and retail trade with several warehouses and stores.

A company owner registers, creates locations, products and roles, and adds employees. Every API call is checked against the employee's own permissions. Stock is tracked per location through an append-only movement ledger.

**Status:** MVP in active development. Auth, users, roles, permissions, categories, products, locations and the inventory core are in place. Purchases, sales, transfers and debts are next.

## Stack

- Go 1.26, [chi](https://github.com/go-chi/chi) router
- PostgreSQL 16, [pgx](https://github.com/jackc/pgx), [goose](https://github.com/pressly/goose) migrations
- Swagger UI ([swag](https://github.com/swaggo/swag))
- Docker Compose for the database, golangci-lint with `depguard` for module boundaries

## Architecture

A modular monolith. Each module is one Go package in `internal/modules/<name>` with `handler → service → repository` layers.

```
cmd/api/              entry point: config, DB, migrations, HTTP server
internal/app/         wiring of modules and the router
internal/modules/     auth, user, role, permission, category, product, location, inventory
internal/platform/    infrastructure without business logic: database, httpx, middleware, health, config
migrations/           goose SQL migrations
api/docs/             generated Swagger spec
```

Key rules:

- A module writes only to its own tables. Other modules see only its exported API.
- Dependencies between modules are fixed in `.golangci.yml`. For example `user → role` is allowed and `role → user` is not.
- `TxManager.WithinTx` puts the transaction into `context.Context`, so an operation that spans several modules stays atomic.
- `inventory_movements` is append-only. A database trigger rejects `UPDATE`, `DELETE` and `TRUNCATE`, and balances can never go below zero.

More detail: [docs/architecture.md](docs/architecture.md) and [docs/api-conventions.md](docs/api-conventions.md) (in Russian).

## Permissions

Permissions use the `object.action` format, for example `products.create` or `roles.assign`. Every protected route is wrapped in a permission middleware.

A role is a package of permissions that is copied to the user on assignment. Changing a role's package updates every user who has that role. A user cannot grant a role or a permission wider than their own. Why it works this way: [ADR 0001](docs/adr/0001-role-permission-packages.md).

## Run locally

Requirements: Go 1.26+, Docker.

```bash
cp .env.example .env
make up      # starts PostgreSQL in Docker
make run     # starts the API on :8080 and applies migrations
```

- Swagger UI: http://localhost:8080/swagger
- Health checks: `GET /healthz`, `GET /readyz`

Start with `POST /auth/register-owner`, then `POST /auth/login`.

## Development

```bash
make test        # go test ./...
make vet
make swag_init   # regenerate Swagger after changing handlers
make migrate-status
```

## API overview

| Area | Endpoints |
|---|---|
| Auth | `POST /auth/register-owner`, `POST /auth/login`, `GET /auth/me`, `POST /auth/logout` |
| Users | `POST /user`, `GET /user/{id}`, `PUT /user/{id}/role` |
| Roles | `GET/POST /role`, `GET/PUT /role/{id}`, `PATCH /role/{id}/archive`, `POST /role/{id}/permission`, `DELETE /role/{id}/permission/{permissionId}` |
| Permissions | `GET /permissions`, `POST /permission/set`, `POST /permission/remove` |
| Catalog | `/category`, `/product` with create, list, get, update and archive |
| Locations | `/location` with create, list, get, update and archive |
| Inventory | `GET /location/{id}/stock`, `GET /location/{id}/stock/{productId}/movements` |

Errors use one format: `{"error": {"code": "...", "message": "..."}}`.

## Roadmap

- Purchases and stock receipts
- Wholesale and retail sales with purchase and sale prices
- Transfers between warehouses and stores
- Sales on credit and debt tracking
