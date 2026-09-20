# 020 — Database integrity and constraints

**Phase:** 4 · **Gate:** G-standard, contributes to G4-data · **Depends on:** 008

## Goal

Fix database-level integrity defects: case-sensitive account links, cascade
deletion of financial history, missing wallet-entry references/indexes, and
inconsistent id types.

## Findings addressed

- Medium: `azeroth_account_links.account_username` uses `text UNIQUE`, so `Bob`
  and `BOB` can both link to the same game account.
- Medium: `store_wallet_entries.user_id` and `store_orders.user_id` use
  `ON DELETE CASCADE`, so deleting a community user erases their order history
  and wallet ledger.
- Low: `store_wallet_entries.order_id`/`actor_id` have no FK or index.
- Low: `azeroth_account_links.account_id` is `integer` while
  `store_orders.account_id` is `bigint`; upstream `account.id` is `int unsigned`.
- Low: `community_users.discord_id` and `discord_identities.discord_id` are both
  `UNIQUE` (redundant, can drift).

## Context

`migrations/00002_identity_discord.sql`, `00003_azeroth_account.sql`,
`00005_store.sql`. This ticket adds a migration. It depends on 008 so the Goose
sequence stays linear.

## Atomic change

Add `00007_db_integrity.sql` with constraint/index changes, plus repository
normalization where needed.

## Requirements

- Enforce case-insensitive account usernames: use `citext`, or a unique index on
  `lower(account_username)`, and normalize on write. Keep the existing API
  contract (usernames remain case-insensitive in AzerothCore).
- Change `store_wallet_entries.user_id` and `store_orders.user_id` to
  `ON DELETE RESTRICT` (or move to soft-delete users); financial rows must not be
  cascaded away.
- Add `REFERENCES store_orders (id)` on `store_wallet_entries.order_id` and an
  index on `order_id` (writes happen in one transaction, so this is safe).
- Make `account_id` consistent (`bigint`) across both tables.
- Remove the redundant uniqueness (keep it in `discord_identities` keyed on
  `(user_id, discord_id)`, or keep `community_users` and document why).
- Down migration restores the previous definitions; data-preserving where
  possible.

## Tests

- Migration `up`/`down` on a scratch database with existing data.
- Test: inserting `Bob`/`BOB` account links conflicts.
- Test: deleting a user with orders/wallet entries is rejected (RESTRICT).
- Test: a wallet entry with an unknown `order_id` is rejected.
- Repository tests for normalization on write.

## Acceptance criteria (gate G-standard, G4-data)

- `task check` green; integration tests pass.
- Financial history cannot be deleted by cascading a user.

## Rollback

Run the down migration. Reverting restores the cascade and the case-sensitive
uniqueness.

## Out of scope

- Soft-deleting users (a separate feature).
- Search/audit indexes (ticket 021).
