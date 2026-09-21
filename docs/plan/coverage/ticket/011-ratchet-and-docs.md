# Ratchet and documentation

**Milestone:** D · **Gate:** G-standard · **Depends on:** 004-010

## Goal

Raise the coverage floor to the achieved total and document the measurement set
so future work cannot silently regress.

## Context

C1 introduces `coverage.floor` at the baseline; the test tickets raise the real
total. Without this ticket the floor stays at the baseline and the ratchet does
nothing.

## Requirements

- Set `coverage.floor` to the achieved measured total (must be ≥ 60%).
- Document in `docs/development.md`: the measured set, the ignore list, how the
  floor is raised, and the `coverage:check` / `web:coverage:check` tasks.
- Note the floor and the SPA threshold in the CI gate description.
- Update the [plan index](../../README.md) and this plan's status table.

## Tests

- `task coverage:check` green at the new floor.
- A deliberate one-line removal from an existing test drops the total below the
  floor and fails the check (verified once, then reverted).

## Acceptance criteria

- Floor ≥ 60% on the measured set; `task check` and CI green.
- Docs describe the set and the ratchet.

## Out of scope

- Further test writing; this ticket only locks in what C4-C10 achieved.
