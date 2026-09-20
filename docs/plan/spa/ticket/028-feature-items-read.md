# Feature: items (read)

## Goal

Browse the item catalog and open an item detail.

## Context

Consumes `GET /api/v1/azeroth/items` and
`GET /api/v1/azeroth/items/{entry}` (ticket 008).

## Requirements

- Route `/items` with filter, class filter, pagination.
- Route `/items/$entry` detail view (renders the `itemview.View` fields).
- Loading/empty/error states.

## Acceptance criteria

- Table and detail render from mocked responses; invalid entry shows the
  unprocessable/404 state.
- `task web:check` green.

## Dependencies

- 008, 020, 022, 023.
