# Demo seed data

## Goal

Idempotent seeds that make the story work on first login.

## Requirements

- Seed the store catalog (`scripts/seed_demo_store.sql`) and the admin role
  (`scripts/seed_demo_admin.sql`) mapped to `ACGW_DEMO_DISCORD_ROLE_ID`.
- `scripts/seed_demo_grant.sql` to fund a wallet by Discord id
  (`task demo:grant`), because the community user exists only after first login.
- Tolerate a missing role id: warn and continue.

## Acceptance criteria

- `task demo:seed` is safe to re-run.
- With the role id set, one login yields both player and staff navigation.

## Dependencies

- 001.
