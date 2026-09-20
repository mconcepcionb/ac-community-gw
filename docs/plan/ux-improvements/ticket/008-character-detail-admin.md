# Admin character detail and mail

## Goal

Make the admin character detail a superset of the player-facing character data,
always reachable by staff, and allow staff to send mail to any character.

## Context

- Detail is served by a shared component used only from the admin route
  (`web/src/routes/admin/characters/$name.tsx`); there is no public character
  page, so "superset of the public page" means the richest data the backend can
  provide.
- Fields today: `guid, name, race, race_name, class, class_name, gender, level,
  online, guild, money, total_time, logout_time`
  (`internal/plugins/azerothcharacter/characters.go:42-56`). No equipment,
  inventory, skills or talents.
- Admin mail is impossible: `handleSendMail` rejects characters not owned by the
  caller's linked account (`internal/plugins/azerothcharacter/mail.go:100-118`).
- Ban/Unban are both always rendered
  (`web/src/features/admin/character-ban-actions.tsx:76-119`).

## Requirements

- Backend: extend the character detail read with equipment (and, if cheap,
  inventory/skills) via the world/character DBs.
- Backend: `POST /api/v1/admin/characters/{name}/mail` under a new
  `azeroth.admin.mail.send` permission that bypasses ownership; reuse the
  existing `SendMailRequest`/validation and audit with action
  `azeroth.admin.mail.send`.
- Console detail:
  - header with quality-independent identity and online/banned badges,
  - equipment panel,
  - **Send mail** action using `MailForm` (with item autocomplete from 011/001),
  - annotations (002) and history (003) panels,
  - state-gated Ban xor Unban.
- Staff can always open the page; the route keeps requiring
  `azeroth.character.list`.

## Acceptance criteria

- Staff can mail a character on any account; the action is audited and confirmed.
- The detail shows equipment and the existing fields without layout regressions.
- Ban/Unban reflects state.
- `task check`, `task openapi:check`, `task web:check` green.

## Implementation notes

- Prefer a single aggregate query for equipment (avoid per-slot queries).
- Do not reuse the self-mail endpoint for admin; a distinct route keeps the
  permission and ownership semantics explicit.
- `MailForm` currently takes free-form character/item ids
  (`web/src/features/characters/mail-form.tsx`); the admin page should pass the
  fixed character and autocomplete item ids.

## Tests

- Go tests for admin mail authorization, ownership bypass and audit.
- RTL + MSW for the detail page and mail flow.

## Dependencies

- 001, 002, 003, 007. Item autocomplete from 011 is optional here.
