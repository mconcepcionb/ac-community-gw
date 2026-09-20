# 022 — Seed fixtures least privilege and idempotency

**Phase:** 4 · **Gate:** G-standard, contributes to G4-data · **Depends on:** 020

## Goal

Make the SQL fixtures and demo seeds safe and repeatable: least privilege, no
accidental full-admin grants, idempotent inserts, and no zeroed credentials.

## Findings addressed

- Medium: `scripts/mysql/init/02_acore_characters.sql` and `03_acore_world.sql`
  grant `ALL PRIVILEGES` on AzerothCore databases to `'acgw'@'%'`, though the
  gateway is documented read-only.
- Medium: `scripts/seed_demo_admin.sql` grants the demo role every permission,
  including all `azeroth.admin.*`/`store.admin.*` scopes.
- Medium: `scripts/seed_demo_store.sql` is not idempotent when the SKU exists
  under a different id (the item insert references hardcoded UUIDs).
- Low: `scripts/mysql/init/01_acore_auth.sql` seeds accounts with zeroed
  salt/verifier.

## Context

`scripts/**`. These run only in development, but fixtures get copied. Ticket 020
may have changed the schema these scripts assume.

## Atomic change

Rewrite the fixtures: `SELECT`-only grants for AzerothCore DBs, explicit admin
scopes, idempotent product/item seeding keyed on the real product id, and
disabled/locked fixture accounts.

## Requirements

- Grant only `SELECT` on `acore_auth`/`acore_characters`/`acore_world` to the
  gateway user; scope the host to the compose network (`'acgw'@'app'`), not `%`.
- `seed_demo_admin.sql`: require an explicit variable/flag to include admin
  scopes, or grant only a documented non-admin subset by default.
- `seed_demo_store.sql`: insert `store_product_items` via
  `SELECT p.id FROM store_products p WHERE p.sku = ...` so it is idempotent under
  `ON CONFLICT (sku) DO UPDATE`.
- Fixture accounts: lock/disable them (or generate real credentials) instead of
  zeroed verifiers; document that they are non-loginable.
- Keep the scripts runnable repeatedly without error.

## Tests

- Run each seed twice against a scratch database and assert no error and no
  duplicate rows.
- Assert the gateway user cannot `INSERT`/`UPDATE` on AzerothCore DBs.
- Assert the demo admin seed without the admin flag does not grant
  `azeroth.admin.*`.

## Acceptance criteria (gate G-standard, G4-data)

- `task check` green; `task db:up` + scripts succeed twice.
- The gateway DB user is read-only on AzerothCore databases.

## Rollback

Revert the scripts; they are development-only.

## Out of scope

- Production migration data.
- Compose credential changes (ticket 029).
