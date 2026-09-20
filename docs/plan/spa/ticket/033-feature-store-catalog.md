# Feature: store catalog

## Goal

Browse and manage store products.

## Context

Consumes product read (ticket 015) and write (ticket 016) operations.

## Requirements

- Route `/store/products`: paginated/searchable table.
- Product detail/edit route or dialog.
- Create/update forms (RHF + Zod) mirroring the product DTO, including
  price/currency and delivery payload validation.
- Delete with confirmation; invalidate the catalog after writes.
- Read-only users (no write permission) see the catalog without actions.

## Acceptance criteria

- CRUD flows work against mocked responses; write actions hidden without
  permission.
- `task web:check` green.

## Dependencies

- 015, 016, 020, 022, 023.
