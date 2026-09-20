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
| Two-surface shell, landing and navigation | implemented |
| Console: moderation, community, catalog/store, delivery, overview | implemented |
| Portal: dashboard, characters, storefront, wallet | implemented |
| Self-service onboarding (create/link and claim) | implemented |
| Retire legacy routes, docs and ADR | implemented |
| Player reports and the unified moderation queue | planned |
| Audit viewer, roles and Discord-role mapping admin | planned |
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
- Reconciliation already exists: `cmd/server/main.go` periodically calls
  `ReconcilePendingOrders`, which completes pending orders whose delivery output
  was persisted after a delivery succeeded.

## Scope

- A two-surface shell: portal layout at `/`, console layout at `/admin`.
- Incremental relocation of every current page into the correct surface, then
  removal of the retired routes.
- Portal features: dashboard, profile, server status, my characters, self-mail,
  storefront, purchase, wallet and orders.
- Console features: live moderation, account ban/GM level, character ban,
  community users and 360 view, item catalog, catalog/store operations,
  character mail delivery, announcements and an operations overview.
- **Now onboarding**: creating and linking a new game account, and claiming an
  existing one with an in-game code.
- Next: player reports, the unified moderation/reconciliation queue, audit
  viewer, roles and Discord-role mapping administration, leaderboards, a public
  no-login surface and API keys.

## Out of scope

- The **Vision** horizon: events calendar, webhooks, embeddable widgets, ops
  metrics dashboard, hosted API docs, public news feed. Roadmap only.
- Redirects or aliases for retired paths: paths are broken and replaced
  (decision in [Migration](#migration)).
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

- **Incremental, always-green**: tickets 001-002 introduce the shell; each
  following ticket relocates one area or builds one feature. The app compiles,
  tests and serves at every step.
- **Retired paths are removed, not redirected** (decision). When an area moves,
  its old route is deleted and the tests and nav that referenced it are updated.
- `task check` (Go) and `task web:check` (Biome + `tsc` + Vitest) must be green
  after every ticket; `task openapi:check` guards contract drift.

### Backend gaps

The Now and Next horizons require these small, well-scoped additions. Each is
delivered inside the vertical slice that needs it, not as a separate track.

| Gap | Needed by |
| --- | --- |
| Self-scoped characters endpoint (ownership enforced) | 009 |
| Self-mail endpoint enforcing character ownership | 009 |
| Per-character public-board opt-in flag | 009, 019 |
| Self-service create/link endpoint | 011 |
| Claim-by-code endpoints and staff claim visibility | 012 |
| Single admin user-360 aggregate endpoint | 004 |
| Player report submission and storage | 014 |
| Store reconciliation read plus manual refund/retry | 015 |
| Audit read capability and permission | 016 |
| Role, permission and Discord-mapping administration | 017 |
| Leaderboard queries per board | 018 |
| Unauthenticated status/boards, short cache, separate rate limit | 019 |
| List-registered-permissions endpoint | 020 |
| API-key authentication and management | 020 |

### Security

- Self-service creation and claiming require **Discord guild membership** and are
  rate-limited per user and per IP; both are audited. One account per user.
- The public surface exposes **full status counts** and **leaderboards with
  character names**, but only for characters whose owner has **opted in**;
  public endpoints get a separate rate limit and a short cache.
- Queue actions on stuck orders (**refund** or **retry**, each with a reason and
  confirmation) are audited; the order state machine still prevents any double
  transition.
- Every mutation is audited; the UI mirrors, but never replaces, backend
  authorization.

### Milestones

| Milestone | Horizon | Tickets |
| --- | --- | --- |
| A - Surfaces | Now | 001-010 |
| B - Onboarding | Now | 011-012 |
| C - Cleanup | Now | 013 |
| D - Staff depth | Next | 014-017 |
| E - Growth and platform | Next | 018-020 |
| Vision | Vision | roadmap only |

### Dependency graph

```
001 -> 002
002 -> 003..010
003..010 -> 013 (retire legacy routes)
008, 009, 010 -> 011 (self-service) -> 012 (claim)
008 -> 014 (player reports) -> 015 (queue)
012 -> 015
004 -> 016 (audit viewer)
002 -> 017 (roles admin)
008, 009 -> 018 (leaderboards) -> 019 (public surface)
017 -> 020 (API keys)
```

001 introduces the route layouts and 002 the landing and navigation; together
they unblock every other ticket. 013 retires the old routes only after 003-010
have moved the areas. Onboarding (011-012) is Now and builds on the portal pages
(008-010). Player reports (014) feed the queue (015), which also consumes account
claims (012).

## Risks

| Risk | Mitigation |
| --- | --- |
| Breaking old paths annoys existing users | Deliberate decision; the shell and landing make the new IA discoverable, and the README documents the removal |
| Two surfaces drift apart in styling | Both consume the same `components/ui` and `components/common` primitives; one shared shell header |
| Player routes leak staff data | Player endpoints are self-scoped server-side and covered by tests; the 360 view uses a dedicated admin aggregate, never the player reads |
| Self-service account creation abused | Discord guild membership plus per-user/per-IP rate limits; audited; one account per user |
| Public names / privacy | Boards show names only for opted-in characters; public endpoints are cached and rate-limited; no account identifiers |
| Manual refund/retry loses money | Confirmation plus a mandatory reason; the SQL state machine prevents double refunds or completing a delivered order |
| Player reports spammed | Reports require an authenticated account and are rate-limited; staff resolve from the queue |
| Leaderboard queries overload the read-only DBs | Paginated queries, cached/refreshed on an interval, individually disableable |
| New backend gaps balloon the vertical slices | Gaps are listed up front and kept minimal; anything larger is pushed to the Vision roadmap |
| Contract drift across many new endpoints | `task openapi:check` in `task check` on every ticket |

## Tickets

### Milestone A - Surfaces (Now)

1. [001-two-surface-layouts.md](ticket/001-two-surface-layouts.md) - portal and console route layouts
2. [002-landing-and-permission-nav.md](ticket/002-landing-and-permission-nav.md) - landing resolver, nav and forbidden state
3. [003-console-moderation.md](ticket/003-console-moderation.md) - live moderation and merged account administration
4. [004-console-community-360.md](ticket/004-console-community-360.md) - community users and the aggregate 360 view
5. [005-console-catalog-store.md](ticket/005-console-catalog-store.md) - item catalog and store operations
6. [006-console-delivery.md](ticket/006-console-delivery.md) - character search, mail delivery and character ban/unban
7. [007-console-overview.md](ticket/007-console-overview.md) - operations overview dashboard
8. [008-portal-home-profile-status.md](ticket/008-portal-home-profile-status.md) - portal dashboard, profile and status
9. [009-portal-characters-mail.md](ticket/009-portal-characters-mail.md) - my characters, self-mail and board opt-in
10. [010-portal-store.md](ticket/010-portal-store.md) - storefront, purchase, wallet and orders

### Milestone B - Onboarding (Now)

11. [011-self-service-account.md](ticket/011-self-service-account.md) - create and link a new game account
12. [012-claim-existing-account.md](ticket/012-claim-existing-account.md) - claim an existing account with an in-game code

### Milestone C - Cleanup (Now)

13. [013-retire-legacy-routes.md](ticket/013-retire-legacy-routes.md) - remove old routes, docs and ADR

### Milestone D - Staff depth (Next)

14. [014-player-reports.md](ticket/014-player-reports.md) - player report submission
15. [015-moderation-queue.md](ticket/015-moderation-queue.md) - unified moderation and reconciliation queue
16. [016-audit-log-viewer.md](ticket/016-audit-log-viewer.md) - audit log viewer and read permission
17. [017-roles-permissions-admin.md](ticket/017-roles-permissions-admin.md) - roles and Discord-role mapping administration

### Milestone E - Growth and platform (Next)

18. [018-leaderboards.md](ticket/018-leaderboards.md) - progression, wealth, playtime and PvP boards
19. [019-public-surface.md](ticket/019-public-surface.md) - no-login status and leaderboards
20. [020-api-keys.md](ticket/020-api-keys.md) - API keys and service accounts

## Vision roadmap

Not ticketed; see [../../use-cases.md](../../use-cases.md) for the stories.

- Events calendar with signups, reminders and rewards (P12).
- Ops metrics dashboard for staff (`/admin/metrics`; distinct from the ticketed
  C5 overview).
- Webhooks / event subscriptions for developers (D3).
- Embeddable status and leaderboard widgets (D4, O4).
- Hosted public API docs and quickstarts (D2).
- Public news feed (O3).

