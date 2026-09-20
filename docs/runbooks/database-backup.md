# Runbook: database backup

## Purpose

Produce a restorable backup of the gateway PostgreSQL database.

## Preconditions

- Access to PostgreSQL.
- `ACGW_DATABASE_URL` or equivalent connection details.
- Sufficient disk space at the destination.

## Procedure

```bash
pg_dump --format=custom --no-owner \
  --file=acgw-$(date +%Y%m%dT%H%M%S).dump \
  "$ACGW_DATABASE_URL"
```

Store the dump in encrypted, access-controlled storage outside the host.

## Validation

```bash
pg_restore --list acgw-<timestamp>.dump >/dev/null && echo OK
```

## Rollback

Not applicable; backups are additive.

## Troubleshooting

- Authentication errors: verify the connection string and role privileges.
- Very large dumps: consider `--compress` and off-peak scheduling.
- Never include secrets in filenames or logs.
