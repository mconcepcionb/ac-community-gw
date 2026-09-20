# Runbook: migrations

## Purpose

Apply, inspect or roll back database migrations.

## Preconditions

- PostgreSQL reachable.
- `ACGW_DATABASE_URL` set (for example in `.env`).
- Goose installed, or the server started with `ACGW_DB_AUTO_MIGRATE=true`.

## Procedure

```bash
task db:status
task db:migrate
# roll back the most recent migration
task db:rollback
```

Always prefer the tasks over invoking Goose directly.

## Validation

```bash
task db:status
curl http://localhost:8080/readyz
```

`readyz` must report the `postgres` check as `ok`.

## Rollback

Roll back one migration with `task db:rollback`. Each migration defines a
`-- +goose Down` section.

## Troubleshooting

- `readyz` returns 503: check `ACGW_DATABASE_URL` and PostgreSQL health.
- Goose cannot find migrations: run from the repository root.
- Dirty migration state: inspect with `task db:status` before retrying.
