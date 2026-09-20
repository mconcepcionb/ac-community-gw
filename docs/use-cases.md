# Use cases

The **desired** use cases for the `ac-community-gw` frontend, organised by
persona and horizon. This is a target-state document: it describes what the
product should let people do, not what today's SPA does. The current page-level
CRUD operations are folded into these use cases as **sub-flows** (see
[Information architecture](#information-architecture)).

## How to read this document

Every use case carries a horizon:

| Horizon | Means |
| --- | --- |
| **Now** | Deliverable with the existing API plus at most small, well-scoped new endpoints. Reorganises today's pages. |
| **Next** | Needs modest backend work (a new query, command or capability) but no architectural change. |
| **Vision** | Directional; implies larger backend work or a new subsystem. |

Each use case is a **job story** (*When ..., I want to ..., so I can ...*)
followed by **acceptance criteria**. A **Sub-flows** line lists the existing
operations the use case absorbs. Permissions are named where an acceptance
criterion depends on them; they are not exhaustive.

## Personas

| Persona | Surface | Needs |
| --- | --- | --- |
| **Player / community member** | Portal (`/`) | Sign in, get a game account, see characters, shop and spend points. |
| **Community / Discord manager** | Console (`/admin/*`) | Look up members, grant points, announce, manage roles. |
| **Game master / in-game admin** | Console | Live moderation, bans, GM levels, mail delivery, queue work. |
| **Store / finance operator** | Console | Catalog, wallets, orders and delivery reconciliation. |
| **Developer / integrator** | API + console | Service credentials, docs, events, embeds. |
| **Read-only observer / streamer** | Public | Status, rankings and news without an account. |

## Surfaces

Two distinct surfaces share one origin, one session cookie and one backend:

- **Player portal** at `/` - community-facing, personal data, self-service.
- **Staff console** at `/admin/*` - operations, moderation and administration.

Rules:

- After sign-in the portal is the default landing surface. Staff reach the
  console from a persistent link and by direct URL.
- Portal users never see console routes: unauthorised access resolves to a
  forbidden/not-found state, not a redirect loop.
- Console users can always return to the portal.
- Both surfaces are permission-gated; the console is not "logged in = allowed".
- Every mutating use case is audited; the portal never exposes a raw command or
  a generic command endpoint.

## Information architecture

### Player portal

| Route | Purpose | Horizon |
| --- | --- | --- |
| `/login` | Discord sign-in | Now |
| `/` | Personal dashboard: account status, balance, characters, recent orders | Now |
| `/onboarding` | Create a new game account or claim an existing one | Next |
| `/characters`, `/characters/$name` | My characters and detail; self-mail | Now |
| `/store`, `/store/products/$sku` | Storefront and product detail | Now |
| `/wallet` | Points balance and order history | Now |
| `/status` | Server status | Now (public in Next) |
| `/leaderboards`, `/leaderboards/$board` | Rankings | Next |
| `/profile` | Discord profile, roles and permissions | Now |
| `/events`, `/events/$slug` | Event list, signup, reminders | Vision |
| `/news` | Public news feed | Vision |

### Staff console

| Route | Purpose | Horizon |
| --- | --- | --- |
| `/admin` | Operations overview dashboard | Now |
| `/admin/users`, `/admin/users/$userId` | Community users and the 360 view | Now / Next |
| `/admin/accounts` | AzerothCore accounts, credentials, links | Now |
| `/admin/characters` | Character search and mail delivery | Now |
| `/admin/items` | Item catalog for catalog building | Now |
| `/admin/online` | Live online moderation | Now |
| `/admin/announcements` | Announcement composer | Now |
| `/admin/store` | Catalog management | Now |
| `/admin/store/wallets` | Wallet and points operations | Now |
| `/admin/store/orders` | Orders and reconciliation | Now |
| `/admin/moderation` | Unified moderation and reconciliation queue | Next |
| `/admin/roles` | Roles and Discord-role mappings | Next |
| `/admin/audit` | Audit log viewer | Next |
| `/admin/api-clients` | API keys / service accounts | Next |
| `/admin/events` | Event administration | Vision |
| `/admin/metrics` | Operations metrics dashboard | Vision |
| `/admin/webhooks` | Event subscriptions | Vision |

### Public / developer

| Route | Purpose | Horizon |
| --- | --- | --- |
| `/status` | Server status without login | Next |
| `/leaderboards` | Rankings without login | Next |
| `/news` | News feed without login | Vision |
| `/docs` | Hosted API reference and quickstarts | Vision |
| widgets | Embeddable status/leaderboard widgets | Vision |

### Folding today's pages into the new IA

| Today | Becomes |
| --- | --- |
| `/` card grid | Portal dashboard + console overview |
| `/identity/users` | `/admin/users` (list) and `/admin/users/$userId` (360) |
| `/account-links`, `/account-links/$userId` | `/admin/users/$userId` and `/admin/accounts` |
| `/accounts`, `/admin/accounts` | `/admin/accounts` + `/admin/moderation` |
| `/admin/online` | `/admin/online` |
| `/characters`, `/characters/$name` | `/characters` for players; `/admin/characters` for staff |
| `/items`, `/items/$entry` | `/admin/items`; players meet items through the storefront |
| `/store/products`, `/store/products/$sku` | `/store` (players), `/admin/store` (staff) |
| `/store/wallet` | `/wallet` (players), `/admin/store/wallets` + `/admin/store/orders` (staff) |
| `/azeroth/status` | `/status` (portal, later public) and the console overview |
| `/profile` | `/profile` |

## Player / community member

### P1 - Sign in with Discord (Now)

*When* I land on the community portal, *I want to* sign in with my Discord
account, *so I can* reach my personal area without another credential.

**Acceptance criteria**

- Sign-in redirects to Discord and returns to the portal; no token is stored in
  the browser.
- A `401` is treated as an anonymous session, never an error screen.
- After sign-in the user lands on the page they intended, or the dashboard.
- Sign-out revokes the server session and clears the principal.
- Missing Discord configuration degrades explicitly (`503`), never breaks the
  portal shell.

**Sub-flows**: today's `/login`, session menu and `RequireAuth` guard.

### P2 - Create and link a game account during onboarding (Next)

*When* I have no AzerothCore account, *I want to* create one and link it to my
Discord identity in one flow, *so I can* play and shop without asking staff.

**Acceptance criteria**

- After first sign-in, a user with no linked account is presented a choice:
  create a new account, or claim one they already own (P3).
- Creating requires a self-chosen username and password that satisfy AzerothCore
  rules; nothing is auto-generated or shown once.
- The account is created and linked in a single flow, and the user lands on
  their characters.
- One game account per community user: a second create/claim is refused with a
  clear explanation.
- Failures (username taken, weak password, world unavailable) are shown inline
  and are retryable.
- The create and link actions are audited.
- Permission: a new self-service link capability; staff retain
  `azeroth.account.link` for support cases.

**Sub-flows**: replaces the admin-only `POST /api/v1/azeroth/accounts` and
`POST /api/v1/azeroth/account-links` for the player's own account.

### P3 - Claim an existing game account with an in-game code (Next)

*When* I already have a game account, *I want to* prove it is mine and link it,
*so I can* keep my characters and use the store.

**Acceptance criteria**

- Starting a claim sends a one-time code to the account holder through the game.
- The player enters the code in the portal within its validity window; success
  links the account and reveals their characters.
- An incorrect or expired code fails safely and can be retried; the flow is
  rate-limited.
- Claiming an account already linked to another user is rejected.
- Pending and failed claims are visible to staff.
- The claim and link are audited.

**Sub-flows**: replaces staff-performed linking for the player's own account.

### P4 - View my characters (Now)

*When* I am signed in, *I want to* see the characters on my linked account, *so
I can* check their progression and gear.

**Acceptance criteria**

- The list shows only the characters of the signed-in user's linked account.
- Opening a character shows the available detail (level, class/race, gear,
  stats).
- Without a linked account the page routes to onboarding instead of showing an
  empty error.
- Permission: `azeroth.character.list` scoped to the user's own link.

**Sub-flows**: today's `/characters`, `/characters/$name` and the staff
`/identity/users/$userId/characters` view.

### P5 - Mail items or money to my own character (Now)

*When* I am viewing one of my characters, *I want to* send it items or money,
*so I can* receive rewards hands-free.

**Acceptance criteria**

- Only characters belonging to the signed-in user's linked account are
  selectable.
- Subject and body are sanitised; the recipient must be a valid character name.
- Success shows the in-game command result; an upstream failure shows a
  retryable error.
- Every delivery is audited.
- Permission: `azeroth.mail.send`, bounded to owned characters.

**Sub-flows**: today's mail form; the staff `.send` capability remains.

### P6 - Browse the storefront (Now)

*When* I want to spend points, *I want to* browse the product catalog, *so I
can* decide what to buy.

**Acceptance criteria**

- The catalog lists active products with their points price and a reward
  summary.
- Opening a product shows its items (name, quality, stats) and/or money reward,
  with the same rendering as the admin view.
- Inactive products are hidden from players.
- Permission: `store.catalog.read`.

**Sub-flows**: today's `/store/products` and `/store/products/$sku` player view.

### P7 - Buy a product with points (Now)

*When* I have enough points, *I want to* buy a product and receive it in game,
*so I can* get my reward without staff involvement.

**Acceptance criteria**

- Purchase requires a linked account and an owned character.
- Insufficient points return a clear message and debit nothing.
- On success points are debited atomically and the reward is delivered; on
  delivery failure the points are refunded.
- The order appears immediately in the wallet as pending, delivered or failed.
- Permission: `store.purchase`.

**Sub-flows**: today's purchase dialog and `POST /api/v1/store/orders`.

### P8 - View my wallet and orders (Now)

*When* I want to know my balance or what I bought, *I want to* see a wallet
page, *so I can* track spending and delivery.

**Acceptance criteria**

- Balance and the append-only ledger are shown.
- Orders show status, product, character and time; failed orders state that
  points were returned.
- The view is personal; other users' wallets are never reachable.
- Permissions: `store.wallet.read`, `store.orders.read`.

**Sub-flows**: today's `/store/wallet`.

### P9 - Check server status (Now, public in Next)

*When* I arrive, *I want to* see whether the server is online, *so I can* know
if I can play.

**Acceptance criteria**

- Connected players, peak, queue and uptime are shown parsed, not as raw
  command text.
- Now: the view requires sign-in. Next: the same data is available
  unauthenticated on the public surface (O1).
- Permission: `azeroth.info.public.read` while authenticated.

**Sub-flows**: today's `/azeroth/status`.

### P10 - See my rank on the leaderboards (Next)

*When* I am active in the community, *I want to* see where I rank, *so I can*
compare and compete.

**Acceptance criteria**

- Boards for character progression, wealth, playtime/activity and PvP/arena.
- Each board paginates and highlights the signed-in user's own position.
- Data reflects the backing databases and refreshes on a defined interval.
- Board definitions are permission-gated and can be turned off individually.

**Sub-flows**: new; not in today's SPA.

### P11 - Manage my profile (Now)

*When* I want to check what the community knows about me, *I want to* open a
profile page, *so I can* see my identity, roles and permissions.

**Acceptance criteria**

- Shows the Discord profile, community user id, member since, roles and
  permissions.
- Roles and permissions update without a re-login, within the refresh interval.
- Permission: `identity.self.read`.

**Sub-flows**: today's `/profile`.

### P12 - Sign up for a community event (Vision)

*When* an event is announced, *I want to* sign up and be reminded, *so I* do
not miss it.

**Acceptance criteria**

- Events have a schedule, capacity and a signup/withdraw action.
- Reminders are shown in the portal before the event.
- Attendance and rewards are tracked and visible to the attendee.

**Sub-flows**: new subsystem; depends on the events design.

## Community / Discord manager

### C1 - Look up a member and see their whole story (Next)

*When* a member asks for help, *I want to* open one page with everything about
them, *so I* do not jump between five screens.

**Acceptance criteria**

- The 360 view shows the Discord profile, roles, linked account, characters,
  wallet, orders and recent audit entries.
- Search is by Discord id, Discord username or community user id.
- Actions available inline (grant, link/unlink, ban) appear only with the
  matching permission.
- The page links to the underlying use cases rather than duplicating their
  mutations.

**Sub-flows**: folds `/identity/users`, `/account-links`, the wallet view and the
current admin account lookups.

### C2 - Grant points (Now)

*When* a member earns a reward, *I want to* grant points with a reason, *so
their* balance reflects it.

**Acceptance criteria**

- Grant by user id or Discord id, a positive amount and a mandatory reason.
- A ledger entry is recorded and the balance updates immediately.
- Granting to an unknown user is refused.
- The action is audited.
- Permission: `store.admin.wallets`.

**Sub-flows**: today's grant dialog and
`POST /api/v1/store/wallets/grant`.

### C3 - Send an announcement (Now)

*When* something happens in the community, *I want to* broadcast an in-game
announcement, *so players* know.

**Acceptance criteria**

- Compose a message and choose the audience supported by the backend.
- A confirmation step precedes sending.
- The delivery result is shown; the action is audited.
- Permission: `azeroth.admin.announce`.

**Sub-flows**: today's announce form.

### C4 - Manage roles and permission mappings (Next)

*When* the community's needs change, *I want to* map Discord roles to internal
roles and grant permissions, *so* access follows Discord.

**Acceptance criteria**

- CRUD mappings from Discord role id to an internal role.
- Assign permissions to roles.
- Changes reach signed-in users within the refresh interval, without a restart.
- Destructive changes require confirmation and are audited.
- Permission: a new roles-administration permission.

**Sub-flows**: new; today this is environment configuration plus SQL.

### C5 - Get an operations overview (Now)

*When* I start my shift, *I want to* see a dashboard summarising the community,
*so I can* spot problems.

**Acceptance criteria**

- Shows server status, pending moderation/reconciliation items, recent
  announcements and recent store activity.
- Each card links into the relevant use case.
- Figures respect the viewer's permissions (no data the viewer cannot open).

**Sub-flows**: today's home card grid becomes a real dashboard.

## Game master / in-game admin

### G1 - Moderate players live (Now)

*When* a player misbehaves, *I want to* see who is online and kick, mute or
unmute them, *so I can* act immediately.

**Acceptance criteria**

- A live list of connected players with their account and character.
- Kick, mute and unmute are available with a reason and show the command result.
- Actions are audited.
- Permissions: `azeroth.admin.players.read`, `.kick`, `.mute`.

**Sub-flows**: today's `/admin/online`.

### G2 - Ban or unban an account and set GM level (Now)

*When* an account needs an administrative decision, *I want to* ban, unban or
change its GM level, *so I can* enforce it.

**Acceptance criteria**

- Ban and unban an account, optionally as a `timeout`; set GM level by realm.
- The current state is visible before acting; the result is shown.
- Actions are audited.
- Permissions: `azeroth.admin.accounts.ban`, `.gmlevel`, `.read`.

**Sub-flows**: today's `/admin/accounts`, ban and GM-level dialogs.

### G3 - Ban or unban a character (Now)

*When* a single character is the problem, *I want to* ban or unban that
character, *so I can* leave the account intact.

**Acceptance criteria**

- Ban and unban a character by name, with the account context shown.
- The command result is shown and the action is audited.
- Permission: `azeroth.admin.characters.ban`.

**Sub-flows**: the endpoint exists today; surface it in the console (today only
account bans have UI).

### G4 - Deliver items or money to a player (Now)

*When* a player is owed something, *I want to* send items or money in-game, *so
they* receive it directly.

**Acceptance criteria**

- Recipient by character name; items and/or money; subject and body sanitised.
- The command result is shown and the action is audited.
- Permission: `azeroth.mail.send` for staff.

**Sub-flows**: today's character mail form.

### G5 - Work the unified moderation and reconciliation queue (Next)

*When* actions need follow-up, *I want to* see them in one queue, *so* nothing
is dropped.

**Acceptance criteria**

- Aggregates items awaiting action: orders pending reconciliation, failed
  deliveries, pending account claims and recent bans/mutes.
- Each item shows enough context to decide and offers a one-click resolution
  where safe.
- Resolutions are idempotent and audited; resolving an item twice has no second
  effect.
- Filters by type, age and assignee.

**Sub-flows**: surfaces the reconciliation path that today lives inside the
order state machine (store.md).

### G6 - Investigate a player (Next)

*When* a report needs judgement, *I want to* combine the 360 view with the audit
trail, *so I can* decide with evidence.

**Acceptance criteria**

- From the 360 view, open the audit entries for that user (C1, A1).
- The investigation is read-only until an action is explicitly taken.
- The decision and its reason are recorded.

**Sub-flows**: links C1 and A1.

## Store / finance operator

### S1 - Manage the catalog (Now)

*When* the rewards change, *I want to* create and edit products, *so* the
storefront stays current.

**Acceptance criteria**

- Create and edit products as single items or bundles; a single-item product
  defaults its sku, name and description from the item.
- Each item id is validated against the catalog; unknown ids are rejected.
- Products can be deactivated without deleting order history.
- The admin preview renders items exactly like the storefront (P6).
- Permissions: `store.admin.products`; catalog reads use
  `store.catalog.read`.

**Sub-flows**: today's `/store/products` admin view and product form.

### S2 - Operate wallets (Now, adjustments in Next)

*When* a balance must change, *I want to* grant, adjust or refund points with a
reason, *so* the ledger stays correct.

**Acceptance criteria**

- Now: grant points (C2).
- Next: manual debit, correction and refund with a mandatory reason.
- Every change is a ledger entry; balances are derived, never edited in place.
- Actions are audited.
- Permission: `store.admin.wallets`.

**Sub-flows**: today's grant dialog; the wallet view becomes an ops view.

### S3 - Monitor orders and reconcile deliveries (Now, queue in Next)

*When* something goes wrong with a purchase, *I want to* find the order and fix
its state, *so* the player is made whole.

**Acceptance criteria**

- List and filter orders by status, user and date range.
- Failed orders show that the points were refunded; stuck pending orders can be
  reconciled idempotently.
- Reconciliation never double-refunds or refunds a delivered order.
- Permission: `store.orders.read`.

**Sub-flows**: today's order list in the wallet page, broadened into an ops
view; the reconciliation action joins the queue (G5).

## Developer / integrator

### D1 - Authenticate as a service with an API key (Next)

*When* I build a bot or a community site, *I want to* use a scoped service
credential, *so I can* call the API without a human session.

**Acceptance criteria**

- Create an API key scoped to a set of permissions; the secret is shown once.
- Keys can be listed, rotated and revoked; last-used time is visible.
- Requests are authorised by the same permission model as sessions.
- Key lifecycle actions are audited.
- Permission: a new API-client administration permission.

**Sub-flows**: new; today the API is cookie-session only.

### D2 - Read public API docs and quickstarts (Vision)

*When* I integrate, *I want to* read a hosted reference, *so I can* start
quickly.

**Acceptance criteria**

- The reference is generated from the committed OpenAPI spec and versioned.
- Quickstarts cover authentication, the service-account flow (D1) and common
  calls.

**Sub-flows**: extends the committed `api/swagger.yaml`.

### D3 - Subscribe to events via webhooks (Vision)

*When* something happens in the gateway, *I want to* receive it, *so* my system
can react without polling.

**Acceptance criteria**

- Register an endpoint and a set of event types.
- Deliveries are signed; failures retry with backoff and are visible in a
  delivery log.
- Event payloads carry identifiers, never secrets (per architecture.md).

**Sub-flows**: new; builds on the in-process event bus.

### D4 - Embed widgets (Vision)

*When* I run an external community page, *I want to* embed gateway data, *so* I
do not rebuild it.

**Acceptance criteria**

- Status and leaderboard embeds are available for public data only.
- Embeds are themable and do not expose authenticated data.

**Sub-flows**: new; shares the public data source of O1 and O2.

## Read-only observer / streamer

### O1 - View public status without login (Next)

*When* I want to know if the server is up, *I want to* see status without an
account, *so I can* check quickly or on stream.

**Acceptance criteria**

- Online counts, peak, queue and uptime are visible unauthenticated.
- No personal or account data is exposed.
- Public traffic is rate-limited separately from authenticated traffic.

**Sub-flows**: same data as P9, on the public surface.

### O2 - View public leaderboards without login (Next)

*When* I want to showcase rankings, *I want to* see them without an account, *so
I can* share or stream them.

**Acceptance criteria**

- The public boards expose the same rankings as P10, minus per-user highlighting.
- Only display data is exposed; no account identifiers.

**Sub-flows**: public projection of P10.

### O3 - Read a public news feed (Vision)

*When* I follow the community, *I want to* read announcements without an
account, *so I can* stay informed.

**Acceptance criteria**

- Published announcements appear in reverse-chronological order.
- Draft or staff-only announcements never appear publicly.

**Sub-flows**: public projection of C3.

### O4 - Embed a widget (Vision)

*When* I run a community page, *I want to* embed status or rankings, *so* the
page stays live with no work.

**Acceptance criteria**

- Identical to D4; the observer is the natural consumer.

### O5 - Hold read-only staff access (Next)

*When* I am responsible for oversight, *I want to* read the console and audit
log without mutating anything, *so I can* review safely.

**Acceptance criteria**

- A read-only role can open the console and A1 but sees no action controls.
- Every read remains permission-gated; the role grants only read permissions.
- The role cannot escalate itself.

**Sub-flows**: new; a role composition over the console's read permissions.

## Audit use cases

These are shared by every staff persona and are listed once.

### A1 - View the audit log (Next)

*When* I need to know who did what, *I want to* search the audit log, *so I can*
account for an action.

**Acceptance criteria**

- Search and filter by actor, target, action and time range.
- Entries show identifiers and outcomes, never secrets, tokens or passwords.
- Results are permission-gated by a new audit-read permission.
- The viewer is read-only; the log cannot be edited from the UI.

**Sub-flows**: new; today audit is written but not surfaced in the SPA.

## Cross-cutting rules

These apply to every use case above:

- **Audit**: every mutation writes an audit entry; entries never contain
  secrets.
- **Authorization**: the UI mirrors the backend; hidden controls are a
  convenience, never the control. Every route and action is permission-gated at
  the API.
- **Degradation**: a missing integration (Discord, SOAP, a database) disables the
  affected use case with an explicit message and leaves the rest of the app
  working.
- **No generic commands**: the portal and console only call typed,
  purpose-specific endpoints.
- **Personal data**: personal and account data is never exposed on the public
  surface.

## Out of scope

- Server-side rendering and final visual design.
- Replacing the Discord bot or Discord's own features.
- Player-earned points (points come from staff grants in these horizons).
- Character rename or appearance customisation.
- Multi-currency or coupon systems beyond today's catalog.

## Open questions

- AzerothCore username and password policy for self-service creation (P2).
- Delivery mechanism and validity window for the in-game claim code (P3).
- Data availability and refresh cadence for playtime and PvP boards (P10).
- Whether public status and boards share one rate-limit budget (O1, O2).
- How API-key scopes map onto the permission registry (D1).
- Ownership of the events subsystem and its data model (P12).

