# Demo

## Goal

Deliver a reproducible demonstration of the gateway that a **prospective server
owner** can be walked through in about five minutes: a player buys a reward with
points and it is delivered in game, then staff act on it in the console - while
the fake AzerothCore dashboard shows the SOAP/CLI layer the gateway hides.

## Status

| Area | State |
| --- | --- |
| Scope, audience and flow decided | done |
| One-command demo environment (`task demo`) | implemented (D1); end-to-end run pending a responsive Docker daemon |
| Demo seed data (`task demo:seed`, `task demo:grant`) | implemented (D2) |
| Scripted walkthrough (`docs/demo.md`) | implemented (D3) |
| One-command reset (`task demo:reset`) | implemented (D4) |
| Demo preflight validation | implemented (D5) |
| README demo section | implemented (D6) |

## Decisions

- **Audience:** prospective server owners. Outcome-focused, minimal jargon; the
  store and moderation flows carry the story.
- **Format:** a one-command local demo plus a scripted walkthrough document.
- **Login:** a **real Discord application** (no dev auth bypass). The demo
  preflight covers app, guild and demo-role setup.
- **Hero flow:** both. A player purchases and watches in-game delivery, then
  staff moderate, grant, and audit the same activity in the console.
- **Backing services:** the compose PostgreSQL + MariaDB fixture and
  `cmd/fakeazerothcore` (SOAP double with a live command dashboard). No real
  AzerothCore and no real game client.

## Context

Most of the demo machinery already exists:

- **Fake AzerothCore** (`cmd/fakeazerothcore`): SOAP double with a live HTML
  dashboard, SSE command journal, `/state`, `/reset`, and request-id
  correlation with the gateway logs. See
  [../runbooks/fake-azerothcore.md](../runbooks/fake-azerothcore.md).
- **Compose fixtures**: MariaDB seeds accounts (ADMIN/PLAYER/BANNED), characters
  (Thrall/Jaina/Arthas with equipment and a guild) and a small item catalog.
- **Seeds**: `scripts/seed_demo_store.sql` (catalog) and
  `scripts/seed_demo_admin.sql` (permission role, mapped to a Discord role id).
- **Two SPA surfaces**: the player portal and the staff console, with the store
  purchase -> in-game delivery flow, account claim by in-game code,
  leaderboards, audit log, roles matrix and community user 360.

What is missing is the packaging: no single entrypoint, thin narrative data, no
script, and no reset. This plan adds them without changing product behaviour.

## Scope

- A `task demo` entrypoint that brings up the databases, migrates and seeds,
  starts the fake AzerothCore and runs the gateway + SPA.
- Idempotent demo seeds: catalog, the demo admin role mapped to a configured
  Discord role id, and a points grant for the demo player.
- A scripted walkthrough (`docs/demo.md`) with the exact clicks and the
  expected result at each step, plus the Discord preflight.
- A one-command reset between runs.
- A README demo section.

## Out of scope

- Any change to the authentication, authorization or serving model.
- A hosted sandbox or a video; the deliverable is a local, reproducible demo.
- Real AzerothCore or a real game client.
- Product features not already implemented.

## Design

### Preflight (once)

The demo uses real Discord login, so before the first run the operator needs:

1. A Discord application (client id, client secret) with the callback URI(s)
   registered - see [../runbooks/discord-oauth-setup.md](../runbooks/discord-oauth-setup.md).
2. A guild the demo account belongs to (`ACGW_DISCORD_GUILD_ID`).
3. A Discord **demo role** the demo account has, whose id is
   `ACGW_DEMO_DISCORD_ROLE_ID`. The seed maps it to the internal
   `ac-core.admin` role, so one identity can be both player and staff.

The gateway must be reachable at the callback origin; `task dev` (Vite + gateway
on localhost) is the default, and Discord accepts `http://localhost` callbacks.

### `task demo`

```
task demo:setup   # docker compose up -d --wait postgres mariadb
                  # task db:migrate
                  # task demo:seed
task demo         # demo:setup, then fake AzerothCore (background) + task dev
```

`task demo` prints the two URLs the operator opens side by side:

- the SPA at `http://localhost:5173/` (login with Discord);
- the fake AzerothCore dashboard at `http://localhost:7878/`.

Between runs, `task demo:reset` restores the fake server's seed state and
re-applies the idempotent seeds.

### Seed data

`task demo:seed` is idempotent and tolerant of a missing role id:

- `scripts/seed_demo_store.sql` - the three-product catalog (exists).
- `scripts/seed_demo_admin.sql` - the `ac-core.admin` role with every
  permission, mapped to `ACGW_DEMO_DISCORD_ROLE_ID` when set.
- A demo points grant for the logged-in player (by Discord id), so the first
  purchase works without a staff grant.

### Walkthrough

`docs/demo.md` is a numbered script with, for each step, what to click and what
the audience should notice, including the "under the hood" beat: the purchase
becomes an SOAP command visible on the fake-server dashboard, correlated by
request id with the gateway log and the audit entry.

## Tickets

1. [001-demo-environment.md](ticket/001-demo-environment.md) - `task demo` /
   `demo:setup` / `demo:seed` and the scripts.
2. [002-demo-seed-data.md](ticket/002-demo-seed-data.md) - idempotent seeds and
   the demo role mapping.
3. [003-walkthrough.md](ticket/003-walkthrough.md) - `docs/demo.md`.
4. [004-demo-reset.md](ticket/004-demo-reset.md) - `task demo:reset`.
5. [005-preflight.md](ticket/005-preflight.md) - validate the environment with a
   clear message before starting.
6. [006-readme.md](ticket/006-readme.md) - README demo section.

## Risks

| Risk | Mitigation |
| --- | --- |
| Discord setup blocks a non-technical audience | The preflight is a short checklist and the walkthrough starts from a logged-in session |
| The two long-running processes are awkward cross-platform | Scripts ship in `sh` and `ps1`; the fake server is backgrounded and cleaned up on exit |
| Demo data drifts from the product | Seeds are idempotent SQL/psql and live next to the existing seeds |
| The console needs staff permissions | The demo role grants them; the same identity is player and staff |
| AzerothCore read features need the MariaDB DSNs | `demo:setup` verifies them; the walkthrough lists the ones required |
