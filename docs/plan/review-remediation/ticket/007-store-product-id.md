# 007 — Store product ID generation

**Phase:** 2 · **Gate:** G-standard, contributes to G2-correctness · **Depends on:** 002

## Goal

Assign a real UUID to every product created through the API so the second and
subsequent creates succeed.

## Findings addressed

- High: `buildProduct` never sets `domain.Product.ID`, so `CreateProduct` inserts
  the nil UUID. The first create succeeds; every later create violates the
  primary key and is misreported as `409 product_exists`.

## Context

`internal/plugins/azerothstore/handlers.go` builds `domain.Product` without an
ID; `repository/store.go` binds `product.ID` directly. The unit fake generates
its own ID and the integration test supplies an explicit ID, which is why the
tests miss it.

## Atomic change

Generate the product ID at creation time in the repository (single source of
truth), keeping `UpdateProduct` keyed by SKU.

## Requirements

- In `CreateProduct`, set `ID: uuid.New()` before insert (do not trust a
  caller-supplied ID for creates).
- Keep `UpdateProduct` unchanged (SKU-keyed).
- Keep `isUniqueViolation` mapping to `ErrProductExists`, which is now only
  reachable for a real SKU collision.
- Ensure the seed/import path that supplies explicit IDs still works (it uses
  the repository; confirm the generated ID does not clobber provided ones — if it
  does, add an internal create variant for seeding).

## Tests

- Handler test: create two products with different SKUs → both `201`, distinct
  IDs, both retrievable.
- Repository test: `CreateProduct` returns a non-nil, non-zero ID.
- Regression test: creating a duplicate SKU still returns `409 product_exists`.

## Acceptance criteria (gate G-standard, G2-correctness)

- `task check` green.
- Two consecutive creates cannot collide on the primary key.
- The fake store and the integration repository agree on ID semantics.

## Rollback

Revert; the bug returns. Low risk to roll back.

## Out of scope

- Product update semantics (tickets 026, 035).
- Seed data (ticket 022).
