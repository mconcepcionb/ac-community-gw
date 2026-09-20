# Feature: characters (read)

## Goal

Browse characters and open a character detail.

## Context

Consumes `GET /api/v1/azeroth/characters` and
`GET /api/v1/azeroth/characters/{name}` (ticket 006).

## Requirements

- Route `/characters` with a searchable table and pagination.
- Route `/characters/$name` detail view.
- Links from the table row to the detail route.
- Loading/empty/error states.

## Acceptance criteria

- Table and detail render from mocked responses; unknown character shows 404
  state.
- `task web:check` green.

## Dependencies

- 006, 020, 022, 023.
