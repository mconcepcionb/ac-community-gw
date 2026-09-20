# Console UX improvements

## Goal

Turn the staff console from a set of flat, wire-level resource pages into a
coherent administrative tool: single-purpose actions per row, detail pages that
aggregate everything known about an entity, structured live data, autocompletion
instead of free-form identifiers, and a real role/permission matrix.

This plan is a review-and-fix pass over the areas called out by staff:
`/admin/accounts`, `/admin/characters`, `/admin/users/<id>`, `/admin/roles`,
`/admin/online` and `/admin/items`.

## Status

| Area | State |
| --- | --- |
| Review of current console and backend | done |
| Shared console primitives (status badges, row actions, pagination, autocomplete) | planned |
| Admin annotations and per-entity logs | planned |
| Accounts: detail page, claim status, unified moderation controls | planned |
| Characters: global browse, superset detail, admin mail | planned |
| Users: consistent read permission | planned |
| Roles: mappings tab + permission matrix | planned |
| Online: structured list and autocomplete | planned |
| Items: in-game-style tooltips and icons | planned (investigation) |

## Context

The console grew feature by feature (see [../portal/README.md](../portal/README.md)
and [../spa/README.md](../spa/README.md)); several interaction patterns were
copy-pasted and the backend never grew the reads the newer UI needs. The review
that produced this plan found:

- **No account detail read.** There is no `GET /api/v1/azeroth/accounts/{username}`
  route; `AccountReader.FindAccountByUsername` exists but is internal-only
  (`internal/core/azerothdb/accounts.go:46`).
- **No ownership signal on accounts.** The list DTO
  (`internal/plugins/azerothaccount/accounts.go:20-31`) has no link/claim field;
  `azeroth_account_links` and `account_claims` are separate tables.
- **No annotation concept** for any entity anywhere in the backend or migrations.
- **Audit cannot be filtered by target.** `GET /api/v1/admin/audit` filters by
  `actor`, `target` (substring only) and `action`; there is no `target_type`
  filter and the stored `metadata` is never returned
  (`internal/plugins/azerothadmin/auditview.go:16-26`).
- **Character listing requires an account** (`missing_account`, 422) so there is
  no global browse; the SPA hides the table until an account is entered
  (`web/src/features/characters/characters-page.tsx:92-133`).
- **Admin mail does not exist.** `POST /api/v1/azeroth/mail` enforces that the
  character belongs to the caller's linked account
  (`internal/plugins/azerothcharacter/mail.go:100-118`).
- **User list and user detail use different permissions** (`identity.user.list`
  vs `azeroth.admin.users.read`); the default demo admin role excludes all
  `azeroth.admin.*`, so it can list users but not open one
  (`scripts/seed_demo_admin.sql:77-81`).
- **The roles endpoint cannot power a matrix.** `GET /api/v1/admin/roles` returns
  only existing grants and mappings — no roles list, no permission catalog
  (`internal/plugins/identitydiscord/rolesadmin.go:29-32`). The catalog lives at
  `GET /api/v1/admin/permissions`, gated by `apikeys.manage`.
- **`.account onlinelist` requires GM security level 4 (SEC_CONSOLE).** Below
  level 4 AzerothCore answers "Command '.account onlinelist' does not exist"
  (`internal/plugins/azerothadmin/moderation.go:53-56`).
- **Items have no icon and no tooltip.** The item view exposes `display_id`
  only, and the backend serves no images/media (`internal/core/itemview/itemview.go:46-76`).

Patterns repeated across pages (fixed by ticket 001) include: both the positive
and negative action rendered unconditionally, `yes/no` text instead of status
badges, duplicated offset-pagination markup with no totals, and free-text
identifier inputs where a lookup should exist.

## Scope

- A shared set of console primitives: status badges, contextual row actions, a
  reusable pagination strip, and an async autocomplete input.
- Generic admin annotations plus audit targeting, so detail pages can show
  staff notes and per-entity history.
- Account detail page, account claim/link status, and unified per-row moderation
  controls.
- Global character browsing with totals, a superset admin character detail, and
  admin-initiated mail.
- Consistent permission for reading community users.
- Roles page split into Discord mappings and a role/permission matrix with batch
  save.
- Structured online list with autocomplete-backed moderation.
- Item tooltips (in-game style) and icons, including the data-source
  investigation.

## Out of scope

- A visual redesign or a new design system; this reuses `components/ui` and
  `components/common`.
- Character equipment/talent editors or any write to AzerothCore beyond the
  existing commands.
- Changing the session, serving or RBAC propagation models (permission
  *naming/ownership* changes are in scope, the mechanism is not).
- The Vision roadmap from the portal plan.

## Design

### Rules

- **One action per state.** A row shows the action that applies now: `Ban` when
  not banned, `Unban` when banned; an enabled/disabled moderation action follows
  the entity's state, not the presence of a name.
- **Reads are aggregated.** A detail page fetches one aggregate read plus
  annotations and audit; it never stitches many list reads client-side.
- **Identifiers are chosen, not typed.** User, account, character and item
  references resolve through the shared autocomplete component.
- **Every mutation stays backend-authorized and audited.** The UI mirrors, never
  replaces, authorization.
- **Namespaces are explicit.** Permissions and routes are namespaced `gw.*`
  (gateway-generic) or `<game>.*` (game-specific); the permission registry
  enforces the prefix so the boundary cannot drift as games are added
  ([ADR 0014](../../ADR/0014-gateway-and-game-permission-namespaces.md)).
- **Backend gaps are delivered inside the vertical slice that needs them.**

### Backend gaps

| Gap | Needed by |
| --- | --- |
| Generic `admin_annotations` table + read/write endpoints (`gw.notes.manage`) | 002 |
| `target_type` audit filter and exposed `metadata` | 003 |
| `gw.`/game permission namespace split; `gw.identity.user.read`; permission catalog for role admins | 004 |
| `GET /accounts/{username}` with link/claim status; owned-by on the list | 005 |
| Optional account on the character list; `COUNT(*)` totals | 007 |
| Admin character mail endpoint; equipment on character detail | 008 |
| `GET /admin/roles` returning roles + permission catalog; batch grant save | 009 |
| Structured online-player read (character DB) | 010 |
| `display_id` to icon mapping and/or self-hosted icon source | 011 |

### Milestones

| Milestone | Tickets |
| --- | --- |
| A - Foundations | 001-004 |
| B - Accounts | 005-006 |
| C - Characters | 007-008 |
| D - Roles | 009 |
| E - Online and items | 010-011 |

### Dependency graph

```
001 -> 005, 006, 007, 008, 009, 010, 011
002, 003 -> 005 (account detail notes/log)
004 -> 005 (users consistency), 009 (catalog)
005 -> 006 (detail links)
007 -> 010 (autocomplete source)
007, 008 -> 008
```

001 unblocks every UI ticket. 002/003 give detail pages their notes and history.
004 settles the user-read permission and the permission catalog before the
account detail and roles tickets consume them.

## Risks

| Risk | Mitigation |
| --- | --- |
| Permission rename breaks existing grants | Hard renames are acceptable pre-production; 004 removes the old definitions and updates every reference (routes, seeds, SPA gating, docs) in one change |
| Permission namespace drifts as games are added | ADR 0014 defines the `gw.`/game-id split and the registry rejects a definition whose name does not match its declared `Namespace` |
| Annotation table becomes a second audit log | Keep annotations authored/edited/deleted by staff; record every change in the audit log |
| Admin mail is abused | Dedicated `azeroth.admin.mail.send` permission, confirmation, reason, audit |
| Global character browse is expensive | Server-side pagination with `COUNT(*)`, indexed name/account filters, conservative default limit |
| Icon source for private-server items | Tooltip renders from data we already own; icons are an optional enhancement with a fallback to no icon (011) |
| Wowhead widget depends on external network and CSP | Investigate only; prefer a self-contained tooltip so custom items render correctly |

## Tickets

### Milestone A - Foundations

1. [001-shared-console-primitives.md](ticket/001-shared-console-primitives.md) - status badges, row actions, pagination, autocomplete
2. [002-admin-annotations.md](ticket/002-admin-annotations.md) - generic staff annotations
3. [003-audit-targeting.md](ticket/003-audit-targeting.md) - target-type filter and metadata
4. [004-permission-namespaces.md](ticket/004-permission-namespaces.md) - `gw.`/game namespace split, user-read permission, catalog

### Milestone B - Accounts

5. [005-account-detail.md](ticket/005-account-detail.md) - account detail endpoint, page, claim status
6. [006-account-list-moderation.md](ticket/006-account-list-moderation.md) - unified GM and ban controls

### Milestone C - Characters

7. [007-character-browse.md](ticket/007-character-browse.md) - global list, totals, filters
8. [008-character-detail-admin.md](ticket/008-character-detail-admin.md) - superset detail and admin mail

### Milestone D - Roles

9. [009-roles-matrix.md](ticket/009-roles-matrix.md) - mappings tab and permission matrix

### Milestone E - Online and items

10. [010-online-structured.md](ticket/010-online-structured.md) - structured online list and autocomplete
11. [011-item-tooltips.md](ticket/011-item-tooltips.md) - in-game tooltips and icons
