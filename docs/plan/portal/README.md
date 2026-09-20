# Portal and console

## Goal

Turn the current flat resource console into two distinct, permission-gated
surfaces - a player-facing **portal** at `/` and a staff **console** at
`/admin/*` - and deliver the desired use cases in
[../../use-cases.md](../../use-cases.md) through the **Now** and **Next**
horizons in incremental, always-green vertical slices.

## Status

| Area | State |
| --- | --- |
| Two-surface shell, landing and navigation | planned |
| Console: moderation, community, catalog/store, delivery | planned |
| Portal: dashboard, characters, storefront, wallet | planned |
| Retire legacy routes, docs and ADR | planned |
| Self-service account create/link and claim | planned |
| Console: queue, audit viewer, roles admin | planned |
| Leaderboards, public surface, API keys | planned |
| Vision horizon (events, webhooks, widgets, metrics, hosted docs) | roadmap |

## Context

- [../../use-cases.md](../../use-cases.md) is the source of truth for the
  desired behaviour this plan implements; each ticket maps to use case ids.
- This plan **supersedes [../spa/README.md](../spa/README.md)** for the frontend.
  The SPA plan remains the historical record of how the current app was built.
- The serving model is unchanged: Caddy serves the built SPA and reverse-proxies
  the API on one origin ([ADR 0012](../../ADR/0012-decoupled-spa-serving.md)).
- The current SPA is a flat list of resource pages at `/accounts`,
  `/account-links`, `/characters`, `/items`, `/identity/users`, `/admin/*`,
  `/store/*`, `/profile`, gated by the `/me` effective permissions.

## Scope

- A two-surface shell: portal layout at `/`, console layout at `/admin`.
- Incremental relocation of every current page into the correct surface, then
  removal of the retired routes.
- Portal features: dashboard, profile, server status, my characters, self-mail,
  storefront, purchase, wallet and orders.
- Console features: live moderation, account ban/GM level, character ban,
  community users and 360 view, item catalog, catalog/store operations,
  character mail delivery, announcements.
- Next: self-service account creation and linking, claiming an existing account
  with an in-game code, unified moderation/reconciliation queue, audit log
  viewer, roles and Discord-role mapping administration, leaderboards, a public
  no-login surface and API keys.

## Out of scope

- The **Vision** horizon: events calendar, webhooks, embeddable widgets, ops
  metrics dashboard, hosted API docs. These are listed as a roadmap, not
  ticketed.
- Redirects or aliases for retired paths: paths are broken and replaced
  (decision in [Design](#migration)).
- Player-earned points; points remain staff grants.
- Character rename/customisation and a visual redesign beyond what the surfaces
  require.
- Any change to the session, RBAC or serving models.

## Design

### Surfaces and landing

- Two layouts share one origin, one session cookie and one backend.
- `/` is the portal shell; `/admin` is the console shell. Console routes are
  siblings under `/admin/*`, never mixed into the portal nav.
- After sign-in, a user holding any console permission lands on `/admin`,
  otherwise on `/`. Console users can always return to the portal.
- Navigation is built from `/me` permissions. A user without the permission sees
  neither the link nor the route (the route resolves to a forbidden state).

### Migration

- **Incremental, always-green**: ticket 001 introduces the shell; each following
  ticket relocates one area or builds one portal feature. The app compiles,
  tests and serves at every step.
- **Retired paths are removed, not redirected** (decision). When an area moves,
  its old route is deleted and the tests and nav that referenced it are updated.
- `task check` (Go) and `task web:check` (Biome + `tsc` + Vitest) must be green
  after every ticket; `task openapi:check` guards contract drift.
- **Onboarding is a Next horizon.** The Now portal tickets (006-008) reference
  `/onboarding`, which lands in 010. Until then an unlinked user sees a graceful
  "account linking is coming" state, never a broken route or an empty screen;
  010 replaces that placeholder.

### Backend gaps

The Now and Next horizons require these small, well-scoped additions. Each is
delivered inside the vertical slice that needs it, not as a separate track.

| Gap | Needed by |
| --- | --- |
| Self-scoped characters endpoint (ownership enforced) | 007 |
| Self-mail endpoint enforcing character ownership | 007 |
| Self-service create/link endpoint | 010 |
| Claim-by-code endpoints and staff claim visibility | 011 |
| Store reconciliation exposure for the queue | 012 |
| Audit read capability and permission | 013 |
| Role, permission and Discord-mapping administration | 014 |
| Leaderboard queries per board | 015 |
| Unauthenticated status/boards with a separate rate limit | 016 |
| API-key authentication and management | 017 |

### Security

- Self-service creation and claiming require **Discord guild membership** and are
  rate-limited per user and per IP; both are audited. One game account per
  community user.
- The public surface exposes display data only, never account identifiers or
  personal data.
- Every mutation is audited; the UI mirrors, but never replaces, backend
  authorization.

### Milestones

| Milestone | Horizon | Tickets |
| --- | --- | --- |
| A - Surfaces | Now | 001-009 |
| B - Self-service onboarding | Next | 010-011 |
| C - Staff depth | Next | 012-014 |
| D - Growth and platform | Next | 015-017 |
| Vision | Vision | roadmap only |

### Dependency graph

```
001 ─┬─ 002 ─ 009
     ├─ 003 ─ 009
     ├─ 004 ─ 009 ── 012
     ├─ 005 ─ 009
     ├─ 006 ─ 009
     ├─ 007 ─ 009
     ├─ 008 ─ 009
     │
     ├─ 010 ─ 011
     ├─ 013
     ├─ 014
     └─ 015 ─ 016
        └─ 017 (independent of 015)
```

001 (the shell) unblocks everything. 009 depends on 002-008 having moved the
areas. 016 depends on 015 (leaderboards) and 006 (status). 010 precedes 011
because both extend onboarding.

The graph omits a few secondary edges for readability: 012 also consumes pending
claims from 011, 013 links from the 360 view of 003, and 016 reuses the status
view from 006.

## Risks

| Risk | Mitigation |
| --- | --- |
| Breaking old paths annoys existing users | Deliberate decision; the shell and landing make the new IA discoverable, and the README documents the removal |
| Two surfaces drift apart in styling | Both consume the same `components/ui` and `components/common` primitives; one shared shell header |
| Permission scoping for players leaks staff data | Player endpoints are self-scoped server-side and covered by tests; portal never calls staff-only endpoints |
| Self-service account creation abused | Discord guild membership plus per-user/per-IP rate limits; audited; one account per user |
| New backend gaps balloon the vertical slices | Gaps are listed up front and kept minimal; anything larger is pushed to the Vision roadmap |
| Leaderboard queries overload the read-only DBs | Paginated queries, cached/refreshed on an interval, individually disableable |
| Contract drift across many new endpoints | `task openapi:check` in `task check` on every ticket |

## Tickets

### Milestone A - Surfaces (Now)

1. [001-two-surface-shell.md](ticket/001-two-surface-shell.md) - portal and console shells, landing and navigation
2. [002-console-moderation.md](ticket/002-console-moderation.md) - live moderation, account bans/GM level, character bans
3. [003-console-community-360.md](ticket/003-console-community-360.md) - community users, links and the 360 view base
4. [004-console-catalog-store.md](ticket/004-console-catalog-store.md) - item catalog and store operations
5. [005-console-delivery.md](ticket/005-console-delivery.md) - character search and mail delivery
6. [006-portal-home-profile-status.md](ticket/006-portal-home-profile-status.md) - portal dashboard, profile and status
7. [007-portal-characters-mail.md](ticket/007-portal-characters-mail.md) - my characters and self-mail
8. [008-portal-store.md](ticket/008-portal-store.md) - storefront, purchase, wallet and orders
9. [009-retire-legacy-routes.md](ticket/009-retire-legacy-routes.md) - remove old routes, docs and ADR

### Milestone B - Self-service onboarding (Next)

10. [010-self-service-account.md](ticket/010-self-service-account.md) - create and link a new game account
11. [011-claim-existing-account.md](ticket/011-claim-existing-account.md) - claim an existing account with an in-game code

### Milestone C - Staff depth (Next)

12. [012-moderation-queue.md](ticket/012-moderation-queue.md) - unified moderation and reconciliation queue
13. [013-audit-log-viewer.md](ticket/013-audit-log-viewer.md) - audit log viewer and read permission
14. [014-roles-permissions-admin.md](ticket/014-roles-permissions-admin.md) - roles and Discord-role mapping administration

### Milestone D - Growth and platform (Next)

15. [015-leaderboards.md](ticket/015-leaderboards.md) - progression, wealth, playtime and PvP boards
16. [016-public-surface.md](ticket/016-public-surface.md) - no-login status and leaderboards
17. [017-api-keys.md](ticket/017-api-keys.md) - API keys and service accounts

## Vision roadmap

Not ticketed; see [../../use-cases.md](../../use-cases.md) for the stories.

- Events calendar with signups, reminders and rewards (P12).
- Ops metrics dashboard for staff (`/admin/metrics`; distinct from the ticketed
  C5 overview).
- Webhooks / event subscriptions for developers (D3).
- Embeddable status and leaderboard widgets (D4, O4).
- Hosted public API docs and quickstarts (D2).
- Public news feed (O3).

