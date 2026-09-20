# Game route namespaces under /admin

## Goal

Move game-specific routes that currently sit under the generic `/api/v1/admin/*`
namespace to the game prefix.

## Context

Two game-specific routes are registered under `/admin`:

- `GET /api/v1/admin/account-claims` (`internal/plugins/azerothaccount/plugin.go:148`).
- `POST /api/v1/admin/characters/{name}/mail` (`internal/plugins/azerothcharacter/plugin.go:99`).

Per ADR 0014 a game route lives under `/api/v1/<game>/*`; game admin actions live
under `/api/v1/<game>/admin/*`.

## Requirements

- Move account claims to `GET /api/v1/azeroth/admin/account-claims`.
- Move character mail to `POST /api/v1/azeroth/admin/characters/{name}/mail`.
- Keep the owning plugins (`azeroth-account`, `azeroth-character`) and their
  permissions (`azeroth.admin.claims.read`, `azeroth.admin.mail.send`).
- Update the SPA call sites:
  - `azerothAdminAccountClaimsList` in `admin-moderation-page.tsx`.
  - `azerothAdminCharactersMail` in `admin-mail-dialog.tsx`.
- Regenerate the client; keep the existing `@ID`s (`azeroth.admin.account_claims.list`,
  `azeroth.admin.characters.mail`) so only the path changes.

## Acceptance criteria

- No game-specific route remains under `/api/v1/admin/*`.
- The moderation page and the admin mail dialog work against the new paths.
- `task check`, `task openapi:check`, `task web:check` green.

## Implementation notes

- Hard move; pre-production, no aliases. Update every reference in the same
  change.

## Tests

- Update the Go handler tests that build the request path and the RTL/MSW tests
  that mock the URL.

## Dependencies

- Independent. Relates to [../../ux-improvements/ticket/008-character-detail-admin.md](../../ux-improvements/ticket/008-character-detail-admin.md).
