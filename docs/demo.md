# Demo

A five-minute, reproducible demonstration of the gateway for prospective server
owners: a player buys a reward with points and it is delivered in game, then
staff act on it in the console - while the fake AzerothCore dashboard shows the
SOAP/CLI layer the gateway hides.

Everything runs locally. The gateway, the SPA, PostgreSQL, RBAC, the audit log
and the store are the real implementations; AzerothCore is a protocol-accurate
double ([fake-azerothcore.md](runbooks/fake-azerothcore.md)), so no worldserver
and no game client are needed.

## What you'll see

- The **player portal**: sign-in with Discord, characters, points, storefront,
  purchase, wallet and orders.
- The **staff console**: live moderation, community user 360, account and
  character administration, store administration, roles and the audit log.
- The **hidden layer**: every action becomes an AzerothCore SOAP/CLI command on
  the fake server's live dashboard, correlated by request id with the gateway
  log and the audit entry.

## Preflight (once)

Requirements: Go 1.27+, Task, Docker + Compose, Node 22+ and pnpm.

1. Configure a Discord application and a guild and create a **demo role** the
   demo account has. Follow
   [runbooks/discord-oauth-setup.md](runbooks/discord-oauth-setup.md) and set in
   `.env`:

   ```dotenv
   ACGW_DISCORD_CLIENT_ID=<application id>
   ACGW_DISCORD_CLIENT_SECRET=<client secret>
   ACGW_DISCORD_REDIRECT_URL=http://localhost:5173/api/v1/auth/discord/callback
   ACGW_DISCORD_GUILD_ID=<guild id>
   ACGW_DEMO_DISCORD_ROLE_ID=<demo role id>   # grants staff permissions to the demo account
   ACGW_DEMO_DISCORD_ID=<your Discord user id>
   ```

2. Confirm the gateway points at the fake AzerothCore and the MariaDB fixture
   (the `.env.example` development defaults):

   ```dotenv
   ACGW_AZEROTH_SOAP_URL=http://localhost:7878/
   ACGW_AZEROTH_LOGIN_DB_DSN=acgw:<MARIADB_PASSWORD>@tcp(localhost:3306)/acore_auth?parseTime=true
   ACGW_AZEROTH_CHARACTER_DB_DSN=acgw:<MARIADB_PASSWORD>@tcp(localhost:3306)/acore_characters?parseTime=true
   ACGW_AZEROTH_WORLD_DB_DSN=acgw:<MARIADB_PASSWORD>@tcp(localhost:3306)/acore_world?parseTime=true
   ```

With the demo role mapped, one identity is both player and staff, so the whole
story can be shown from one login.

## Start

```bash
task demo
```

`task demo` starts the databases, migrates and seeds, launches the fake
AzerothCore in the background, and runs the gateway and the SPA together. Open
two windows side by side:

- the SPA at <http://localhost:5173/>;
- the fake AzerothCore dashboard at <http://localhost:7878/>.

## The script

### 1. Sign in (30 s)

Open the SPA and click **Login with Discord**. After the callback you land on the
portal dashboard.

Notice: the landing page and navigation are driven by your effective
permissions; the console link appears because the demo role grants staff
permissions.

### 2. Player: buy a reward and watch it deliver (2 min)

1. **My characters** - Thrall and Jaina are read live from the AzerothCore
   character database.
2. Fund the wallet (first run only), then refresh the wallet page:

   ```bash
   task demo:grant            # adds 1000 points to ACGW_DEMO_DISCORD_ID
   ```

3. **Store** - buy **Traveler's Backpack** for Thrall.
4. Switch to the fake AzerothCore dashboard. The purchase appears as a SOAP
   `.send items` command, with the same request id you can find in the gateway
   log.
5. Back in the SPA, **Wallet / Orders** shows the order as delivered and the
   points debited.

Notice: the SPA never mentions SOAP, CLI syntax or database schemas; the gateway
translated an authorized REST call into a validated in-game command.

### 3. Staff: act on the same activity (2 min)

1. **Console -> Overview**: the operations dashboard.
2. **Community users -> a user 360**: identity, linked account, characters and
   store activity in one view.
3. **Live moderation**: kick, mute or announce; watch each command land on the
   fake AzerothCore dashboard in real time.
4. **Store administration**: grant points (the ledger entry appears in the
   wallet), list orders, and refund or retry a stuck order.
5. **Roles**: the role/permission matrix and the Discord-role mappings.
6. **Audit**: every purchase and moderation action is recorded; open a user or
   account to see its per-entity history.

### 4. The punchline (30 s)

The dashboard and the audit log tell the same story from two sides: a single
request id links the SPA action, the gateway log, the AzerothCore command and
the audit row. The administrative surface projected to consumers is
explicit, authorized and auditable operations - never a raw command endpoint.

## Reset between runs

```bash
task demo:reset   # clears the fake AzerothCore state and re-applies the seeds
```

## Troubleshooting

| Symptom | Fix |
| --- | --- |
| Login fails or returns an error | Check the Discord app, redirect URI and guild; see [discord-oauth-setup.md](runbooks/discord-oauth-setup.md) |
| No console navigation after login | `ACGW_DEMO_DISCORD_ROLE_ID` must be mapped and the gateway restarted so grants load; log out and back in |
| No characters or accounts | Set the three `ACGW_AZEROTH_*_DB_DSN` variables to the compose MariaDB |
| Purchase fails with "link an account first" | Claim or create a game account from the portal onboarding page first |
| Dashboard shows no commands | The gateway must point at `ACGW_AZEROTH_SOAP_URL=http://localhost:7878/` |

## See also

- [Demo plan](plan/demo/README.md)
- [fake-azerothcore.md](runbooks/fake-azerothcore.md)
- [use-cases.md](use-cases.md)
