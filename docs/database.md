# Database

## Two different databases

- **PostgreSQL** is the gateway's own database.
- **MySQL/MariaDB** databases (`auth`, `characters`, `world`) belong to
  AzerothCore and are external and read-only from the gateway's perspective.

Do not confuse them.

## Access layer

```
database/sql + pgx driver
sqlc  -> typed generated access
Goose -> migrations
```

No ORM is used. SQL stays explicit.

## Ownership

Tables have a conceptual owner even though the schema lifecycle is global.

| Owner | Tables |
| --- | --- |
| core | `roles`, `permissions`, `role_permissions`, `audit_log` |
| identity-discord | `community_users`, `discord_identities`, `sessions` |
| azeroth-account | `azeroth_account_links` |
| azeroth-store (planned) | `wallets`, `ledger_entries`, `products`, `orders` |

## Migrations

A single global Goose sequence lives in `migrations/`:

```
migrations/
    00001_core.sql
    00002_identity_discord.sql
    00003_azeroth_account.sql
    ...
```

Migrations are SQL. Each file has `-- +goose Up` and `-- +goose Down`.

```bash
task db:migrate
task db:status
task db:rollback
```

The migration files are embedded into the binary (`migrations/embed.go`) so the
server can optionally apply them at startup with `ACGW_DB_AUTO_MIGRATE=true`.

## sqlc

Queries are owned per module and live next to the repository that uses them:

```
internal/plugins/identitydiscord/repository/queries.sql
internal/plugins/identitydiscord/repository/generated/
```

`sqlc.yaml` uses the `migrations/` directory as the schema source, so the
schema is never duplicated by hand. Generated code is committed.

```bash
task codegen:sqlc
```

## Integrity and indexes

- Account usernames are case-insensitive (unique index on
  `lower(account_username)`), matching AzerothCore.
- Financial history is protected: `store_wallets`, `store_wallet_entries` and
  `store_orders` reference `community_users` with `ON DELETE RESTRICT`, so
  deleting a user cannot cascade away their orders or ledger.
- `store_wallet_entries.order_id` references `store_orders`; the order status and
  price columns have `CHECK` constraints.
- User search uses GIN trigram indexes; audit filters (`actor_id`, `action`,
  `target_type`) and a BRIN index on `occurred_at` support retention pruning.

## See also

- [ADR 0003](ADR/0003-postgresql-sqlc-goose.md)
- [azerothcore-integration.md](azerothcore-integration.md)
