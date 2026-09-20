# Feature: AzerothCore status

## Goal

Show the AzerothCore server status in the SPA.

## Context

Consumes `GET /api/v1/azeroth/info/status` (ticket 005).

## Requirements

- Route `/azeroth/status` (protected, permission per ticket 005).
- Card/summary view of the parsed status fields.
- Explicit unavailable/error state when the SOAP upstream is down (503/502).

## Acceptance criteria

- Renders status from a mocked response; shows the error state on failure.
- `task web:check` green.

## Dependencies

- 005, 020, 022, 023.
