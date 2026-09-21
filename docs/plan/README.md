# Plans and roadmap

This is the single index of planned work and the consolidated roadmap. It
replaces the per-plan "vision roadmap" and "deferred follow-up" sections that
were scattered across individual plan READMEs: those items now live here, so
there is one backlog.

Docs split (see [../README.md](../README.md)): `docs/*.md` describes the system
**today**, `docs/ADR/` the **why**, `docs/plan/` the **what**, and
`docs/runbooks/` the **how to operate**.

## Plans

A plan is a goal with tickets, milestones, gates and a dependency graph. When a
plan is delivered, its tickets stay as the historical record and its decisions
are consolidated into the stable docs and an ADR.

| Plan | Scope | Status |
| --- | --- | --- |
| [foundation](foundation/README.md) | Skeleton, core registries, HTTP API, PostgreSQL/sqlc, SOAP adapter, plugin system | delivered |
| [discord-auth](discord-auth/README.md) | OAuth, sessions, provisioning, role sync, hardening | delivered |
| [spa](spa/README.md) | OpenAPI pipeline, typed client, app shell, domain features, serving | delivered (historical; superseded by `portal`) |
| [portal](portal/README.md) | Player portal and staff console, onboarding, staff depth, growth | delivered (see note) |
| [ux-improvements](ux-improvements/README.md) | Console primitives, detail pages, roles matrix, online, item tooltips | delivered |
| [gateway-admin](gateway-admin/README.md) | Route/permission ownership split; gateway vs game plugins | delivered |
| [review-remediation](review-remediation/README.md) | 42-ticket code-review remediation (security, correctness, hardening) | delivered |
| [coverage](coverage/README.md) | Coverage tooling, linter, domain/repository tests, ratchet | delivered |
| [secrets](secrets/README.md) | SOPS + age secret management and credential rotation | active |

> **Portal note:** tickets 014-020 (player reports, moderation queue, audit
> viewer, roles admin, leaderboards, public surface, API keys) are implemented
> and committed (`f50cb88`..`c641acc`); the status table in that README is
> pending a refresh.

## Active work

1. **[secrets](secrets/README.md)** (S1-S7) - SOPS + age secret management with
   a dedicated age identity, an encrypted `secrets/development.sops.env`, a
   static leak guard in CI and the `task secrets:*` flow. S1-S5 are implemented;
   **S6 (rotate the exposed Discord/SOAP credentials) and S7 (the production
   secret set on the deploy host) are pending operator actions** - the runbooks
   and `secrets/production.env.example` are in place.

## Delivered

- **[coverage](coverage/README.md)** (C1-C11) - delivered. Measured coverage
  61.4% unit / 76.3% with integration; `coverage.floor` 61.0; SPA 68.4%
  statements with a 60% threshold; `golangci-lint` in the Go gate; a CI
  integration job with Postgres + MariaDB.

## Roadmap

Everything not ticketed by an active plan, coalesced from the delivered plans'
deferred sections, the [status report](../status-report.md) and the
[use cases](../use-cases.md). Ordered by horizon.

### Near term (hardening the MVP)

| Item | Source | Notes |
| --- | --- | --- |
| Run the real gate on `origin` | status report | CI exists (`ci.yml`) and the remote is now configured; confirm `task check`, `test:race`, `test:integration` and `openapi:check` are green on CI |
| Execute the integration suite regularly | review-remediation 033, status report | Postgres + MariaDB services; currently only run locally on demand |
| SPA URL-state re-sync, centralized retry policy, `SameSite=None` CSRF token | review-remediation 038 | Explicit follow-ups recorded when 038 shipped |
| Pin base-image digests | review-remediation 032 | Requires registry access; operational |
| `git`/CI baseline commit for the OpenAPI drift check | review-remediation 001 | Resolved now that the repo has `origin` and history |
| Rotate exposed Discord/SOAP credentials and supply Compose passwords | review-remediation (operational), secrets S5 | Tracked by the [secrets](secrets/README.md) plan |

### Next (architectural follow-ups)

| Item | Source | Notes |
| --- | --- | --- |
| Gateway `public` aggregator plugin | gateway-admin 005 | Game public data already moved under `/azeroth/public/*`; the aggregating `/public/*` plugin was deferred pending a multi-game public model |
| Multi-game 360 response shape | gateway-admin README | Current 360 is flat; a `games`-keyed shape is a separate decision, not needed for the ownership move |
| Store schema for multi-game delivery | gateway-admin 007 | The plugin split shipped; the product/wallet/order schema redesign is a follow-up |
| Item icons | ux-improvements 011 | No `displayid` -> icon source today; tooltip ships without icons |
| Autocomplete in mail/grant forms | ux-improvements 011 | Deferred from the tooltip ticket |

### Vision (not ticketed)

From the [portal plan](portal/README.md) and [use-cases.md](../use-cases.md):

- Events calendar with signups, reminders and rewards.
- Ops metrics dashboard for staff (`/admin/metrics`, distinct from the console
  overview).
- Webhooks / event subscriptions for developers.
- Embeddable status and leaderboard widgets.
- Hosted public API docs and quickstarts.
- Public news feed.

## Rules

1. New work enters as a ticket in a plan (for a goal) or as a roadmap row (not
   yet shaped). Roadmap rows move into a plan with tickets before implementation.
2. A delivered plan is marked up here; its deferred items are lifted into the
   roadmap above and its decisions recorded in an ADR.
3. Plans are numbered and ordered; the execution rules of each plan
   (atomic tickets, gates) apply unless the plan states otherwise.
