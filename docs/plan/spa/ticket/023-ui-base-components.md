# UI base components

## Goal

Provide the reusable component layer the features build on.

## Context

The focus is building blocks, not final layout. Components must be generic and
permission/error aware where useful.

## Requirements

- shadcn/ui primitives initialised: Button, Input, Label, Textarea, Card, Dialog,
  AlertDialog, Table, Badge, Skeleton, Tabs, Select, Switch, Tooltip, Sonner
  (toasts), Form (react-hook-form integration).
- Common components in `src/components/common/`:
  - `DataTable` (TanStack Table wrapper with loading/empty/error states),
  - `EmptyState`, `ErrorState`, `LoadingState`,
  - `ConfirmDialog` (for destructive actions),
  - `MutationButton` (disabled/pending state around a mutation),
  - `PageHeader`,
  - `PermissionGate` (wraps `<Can>` with a fallback),
  - `FormField` helpers wiring RHF + Zod + shadcn `Form`.
- A `cn` utility and design tokens via Tailwind theme.

## Acceptance criteria

- Components are used by at least one placeholder page to prove composition.
- `DataTable` renders rows, loading and empty states.
- `PermissionGate` hides content without the permission.
- `MutationButton` shows pending state and surfaces `ApiError` messages.
- `task web:check` green.

## Implementation notes

- Keep components presentational and free of domain types; features pass data.
- Prefer composition over configuration; avoid speculative props.

## Tests

- `DataTable` (rows/empty/loading), `PermissionGate`, `MutationButton`
  (pending/error).

## Dependencies

- 021, 022 (for permission helper). Unblocks 024..034.
