# Runbook: database restore

## Purpose

Restore the gateway PostgreSQL database from a backup.

## Preconditions

- A validated backup file.
- Target database identified, with a fresh backup of its current state.
- Confirmation that no migration is incompatible with the backup schema version.

## Procedure

```bash
# stop the gateway to avoid writes
pg_restore --clean --if-exists --no-owner \
  --dbname "$ACGW_DATABASE_URL" \
  acgw-<timestamp>.dump
```

Then start the gateway.

## Validation

```bash
task db:status
curl http://localhost:8080/readyz
```

Confirm the schema version and readiness.

## Rollback

Restore the pre-restore backup if the restore is wrong.

## Troubleshooting

- Version mismatch: apply the matching migrations or use a matching backup.
- Locked objects: ensure the gateway and other writers are stopped.
- Never restore into an environment without a pre-restore backup.
