# 040 — Accessibility fixes

**Phase:** 6 · **Gate:** G-standard, contributes to G6-frontend · **Depends on:** 001

## Goal

Make data tables and pagination accessible.

## Findings addressed

- Low: `TableHead` renders `<th>` without `scope`, `DataTable` has no
  `<caption>`/`aria-label`, and pagination controls are not in a `<nav>` and lack
  `aria-label`s.

## Context

`web/src/components/common/table.tsx`,
`web/src/components/common/data-table.tsx`, and their consumers.

## Atomic change

Add table semantics and accessible labels without changing layout.

## Requirements

- `TableHead`: emit `scope="col"` (and `scope="row"` where appropriate).
- `DataTable`: accept a required `caption`/`aria-label`; render a visually hidden
  caption; wrap pagination in `<nav aria-label="Pagination">`.
- Pagination buttons: descriptive `aria-label` (`Previous page`, `Next page`).
- Ensure the sortable headers expose `aria-sort`.
- Pass captions from each feature page.

## Tests

- Component tests assert `scope`, caption/label, `nav`, and `aria-sort`.
- A Playwright/axe pass on a representative table reports no critical violations.

## Acceptance criteria (gate G-standard, G6-frontend)

- `task web:check` green.
- Axe reports no critical table/navigation violations on the main list pages.

## Rollback

Revert; purely additive attributes, low risk.

## Out of scope

- A full WCAG audit of forms/dialogs.
