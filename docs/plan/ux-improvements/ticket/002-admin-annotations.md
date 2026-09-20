# Admin annotations

## Goal

Let staff attach free-text notes ("annotations") to accounts, community users
and characters, visible to other staff.

## Context

There is no note/annotation concept anywhere in the backend or migrations; the
only log is the generic `audit_log` (`migrations/00001_core.sql:20-34`), which
records actions, not comments. Account detail pages are being introduced by 005
and need a place for staff context (e.g. "chargeback risk", "already refunded
manually").

## Requirements

- Migration `admin_annotations`:
  - `id uuid PK`, `target_type text NOT NULL`, `target_id text NOT NULL`,
    `author_id uuid NOT NULL`, `body text NOT NULL`,
    `created_at timestamptz NOT NULL`, `updated_at timestamptz`.
  - Index on `(target_type, target_id, created_at)`.
- Endpoints under `GET/POST/PATCH/DELETE /api/v1/admin/annotations` with
  `?target_type=&target_id=` filtering and `limit`/`offset`.
- Permissions (gateway-generic, ADR 0014): `gw.notes.read` to read,
  `gw.notes.write` to create and edit/delete your own, and `gw.notes.manage` to
  moderate (edit/delete) any note. Authorship is taken from the session, never
  the body.
- Frontend: a reusable `AnnotationsPanel` (list, add, edit, delete) plus hooks,
  built in this ticket and embedded by the detail pages (005, 008).
- Every create/update/delete is itself recorded in the audit log.

## Acceptance criteria

- Annotations round-trip and are returned newest-first per target.
- A user cannot edit or delete another author's note; staff with the manage
  permission can moderate any.
- DTOs are typed and present in `api/swagger.yaml`; `task openapi:check` green.
- Go tests: authorization, authorship enforcement, audit writes.

## Implementation notes

- Follow an existing plugin's repository shape (sqlc under
  `repository/queries.sql` + generated code) rather than raw SQL in handlers.
- `target_type` is a closed set (`account`, `user`, `character`); reject others.
- Prefer this generic table over an accounts-only table so all detail pages
  reuse it.

## Tests

- Go handler tests for validation and authorship/moderation authorization.
- RTL + MSW test for the annotations panel (list, add, write-gated controls).

## Dependencies

- Independent; consumed by 005 (accounts), 008 (characters) and optionally the
  user-360 page.
