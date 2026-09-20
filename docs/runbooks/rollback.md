# Runbook: rollback

## Purpose

Return the gateway to the previous working version.

## Preconditions

- Known previous version (image tag or revision).
- Awareness of whether the release included a migration.

## Procedure

1. Switch the runtime back to the previous version.
2. If the release added a migration and it is safe to reverse, roll it back:

```bash
task db:rollback
```

3. Restart the service.

## Validation

```bash
curl https://<host>/healthz
curl https://<host>/readyz
```

Confirm the plugin/command list matches the expected version.

## Rollback

Re-deploy the version you just rolled back from, if needed.

## Troubleshooting

- If a migration is destructive or irreversible, do not roll the schema back;
  restore from the database backup instead
  (see [database-restore.md](database-restore.md)).
