# AzerothCore integration

## Adapters

AzerothCore is reached through replaceable adapters, all confined to
`internal/adapters/`:

- **SOAP** (implemented): executes administrative commands.
- **Login DB** (implemented): read-only MySQL/MariaDB access used to list
  accounts, which the SOAP interface cannot do.
- **Character DB** (implemented): read-only MySQL/MariaDB access to characters.
- **World DB** (implemented): read-only MySQL/MariaDB access to the item catalog.

Plugins never import adapters. They depend on core contracts, and
`cmd/server` injects the concrete adapter.

## SOAP transport

AzerothCore exposes a very small SOAP interface centred on
`executeCommand(command)`. The transport contract is:

```go
type CommandExecutor interface {
    Execute(ctx context.Context, command string) (string, error)
}
```

`internal/adapters/azerothsoap` implements it with:

- `net/http`
- hand-built SOAP XML (no SOAP library)
- HTTP Basic Auth
- timeouts
- HTTP status and SOAP fault translation

The adapter is **not** the owner of the command catalogue.

## Command ownership

The plugin that owns a capability also owns the AzerothCore syntax required to
implement it.

```
azeroth-account   owns: account create, account set password, account set email
azeroth-admin     owns: ban/unban account, account set gmlevel, kick, mute/unmute,
                        ban/unban character, announce, onlinelist
azeroth-character owns: character reads, send item/mail/money, rename, customize (rename/customize planned)
```

There is no global package containing every AzerothCore command.

## Application command vs AzerothCore command

```
azeroth-store
    |  character.send-item          (application command, typed)
    v
core command registry
    |
    v
owner plugin handler              (builds the AzerothCore CLI string)
    |
    |  .send items ...
    v
SOAP CommandExecutor
```

`azeroth-store` never knows `.send items`; only the owning plugin does.

Example in this iteration:

```
account.change-password  ->  azeroth-account  ->  ".account set password ..."
account.ban              ->  azeroth-admin    ->  ".ban account ..."
```

## HTTP surface

Plugins expose explicit, permission-gated endpoints; there is never a generic
command endpoint (see [security.md](security.md)).

| Method | Path | AzerothCore command | Permission |
| --- | --- | --- | --- |
| `GET` | `/api/v1/azeroth/info/status` | `.server info` | `azeroth.info.public.read` |
| `GET` | `/api/v1/azeroth/accounts` | login DB read (list) | `azeroth.account.list` |
| `POST` | `/api/v1/azeroth/account-links` | gateway DB (link user↔account) | `azeroth.account.link` |
| `GET` | `/api/v1/azeroth/account-links` | gateway DB (list links) | `azeroth.account.read` |
| `GET` | `/api/v1/azeroth/account-links/{user_id}` | gateway DB (read one) | `azeroth.account.read` |
| `DELETE` | `/api/v1/azeroth/account-links/{user_id}` | gateway DB (remove) | `azeroth.account.link` |
| `GET` | `/api/v1/azeroth/characters` | character DB read (list) | `azeroth.character.list` |
| `GET` | `/api/v1/azeroth/characters/{name}` | character DB read (one) | `azeroth.character.list` |
| `GET` | `/api/v1/azeroth/users/{user_id}/characters` | character DB read (linked account) | `azeroth.character.list` |
| `GET` | `/api/v1/azeroth/items` | world DB read (item search) | `azeroth.item.list` |
| `GET` | `/api/v1/azeroth/items/{entry}` | world DB read (one item) | `azeroth.item.list` |
| `POST` | `/api/v1/azeroth/mail` | `.send items` / `.send money` | `azeroth.mail.send` |
| `GET` | `/api/v1/azeroth/online` | `.account onlinelist` | `azeroth.admin.players.read` |
| `POST` | `/api/v1/azeroth/players/{name}/kick` | `.kick` | `azeroth.admin.players.kick` |
| `POST` | `/api/v1/azeroth/players/{name}/mute` | `.mute` | `azeroth.admin.players.mute` |
| `POST` | `/api/v1/azeroth/players/{name}/unmute` | `.unmute` | `azeroth.admin.players.mute` |
| `POST` | `/api/v1/azeroth/characters/{name}/ban` | `.ban character` | `azeroth.admin.characters.ban` |
| `POST` | `/api/v1/azeroth/characters/{name}/unban` | `.unban character` | `azeroth.admin.characters.ban` |
| `POST` | `/api/v1/azeroth/announce` | `.announce` | `azeroth.admin.announce` |
| `POST` | `/api/v1/azeroth/accounts` | `.account create` | `azeroth.account.manage` |
| `POST` | `/api/v1/azeroth/accounts/{username}/password` | `.account set password` | `azeroth.account.manage` |
| `PUT` | `/api/v1/azeroth/accounts/{username}/email` | `.account set email` | `azeroth.account.manage` |
| `PUT` | `/api/v1/azeroth/accounts/{username}/gmlevel` | `.account set gmlevel` | `azeroth.admin.accounts.gmlevel` |
| `POST` | `/api/v1/azeroth/accounts/{username}/ban` | `.ban account` | `azeroth.admin.accounts.ban` |
| `POST` | `/api/v1/azeroth/accounts/{username}/unban` | `.unban account` | `azeroth.admin.accounts.ban` |

`GET /api/v1/azeroth/info/status` returns the parsed online counts (connected
players, characters in world, connection peak, queue, uptime) alongside the raw
`.server info` output, so a portal can render live status without parsing text.

Command endpoints return `{"result": "<AzerothCore output>"}`. Validation failures return
`422 unprocessable_entity` and upstream failures `502 bad_gateway`. The SPA
(`web/`, see [frontend.md](frontend.md)) exposes all of them once logged in.

### User / account links

An administrator links a community user to an AzerothCore account with
`POST /api/v1/azeroth/account-links`, identifying the user by `discord_id`,
`discord_username` or `user_id`:

```json
{"discord_username": "alice", "account_username": "ADMIN"}
```

To find those values, search community users with
`GET /api/v1/identity/users?filter=` (`identity.user.list`) and AzerothCore
accounts with `GET /api/v1/azeroth/accounts?filter=` (`azeroth.account.list`).
A `discord_username` that matches more than one user returns
`409 ambiguous_user` with the candidates, so the admin can retry by id.

The link is stored in the gateway database (`azeroth_account_links`, one account
per user). The account is verified read-only against the login DB (the SOAP
interface cannot), and its id is stored alongside the username. `discord_id`
resolution uses the `identity.user.directory` capability published by
`identity-discord`, so `azeroth-account` never reads another module's tables.

## Login database

`internal/adapters/azerothmysql` reads the AzerothCore login database
(`account`, `account_access`, `account_banned`) using `ACGW_AZEROTH_LOGIN_DB_DSN`.
It is read-only and backs `GET /api/v1/azeroth/accounts`. The fake worldserver
can mirror its account state into the same database (`-login-db-dsn`). The
`compose.yaml` MariaDB fixture reproduces the official base schema
(`data/sql/base/db_auth/*.sql`), so the adapter runs unchanged against a real
server.

If the DSN is unset the endpoint returns `503 login_db_not_configured`; if it is
set but unreachable (at startup or at query time) it returns
`503 login_db_unavailable`. Either way the gateway keeps running: SOAP commands,
authentication and the rest of the API are unaffected.

## Characters and delivery

The same adapter reads the character DB (`ACGW_AZEROTH_CHARACTER_DB_DSN`) and
backs `azeroth-character`. The `acore_characters` fixture also mirrors the
official base schema (`characters`, `guild`, `guild_member`).

- `GET /api/v1/azeroth/characters?account=<username>` lists a login account's
  characters; `/users/{user_id}/characters` resolves the linked account first
  (via the `azeroth.account.directory` capability) and requires no account name.
- `POST /api/v1/azeroth/mail` delivers items and/or money in-game with
  `.send items` / `.send money`. The recipient name is validated (letters only)
  and the subject/body are sanitized, because the values are embedded in the
  command string.

Delivery is a live-world action, so it always goes through SOAP; characters are
read data, so they come from the character DB.

## Item catalog

`azeroth-item` reads `item_template` from the world DB
(`ACGW_AZEROTH_WORLD_DB_DSN`) and exposes a read-only catalog:

```
GET /api/v1/azeroth/items?filter=<name>&class=<id>&limit=&offset=
GET /api/v1/azeroth/items/{entry}
```

Each item returns `entry`, `name`, `class`/`class_name`, `subclass`, `quality`/
`quality_name`, `item_level`, `required_level`, `inventory_type`, `buy_price`,
`sell_price` and more. Store products reference item entries, so operators use
this endpoint (or the demo's *item search*) to pick ids for the catalog.

## Testing without AzerothCore

- The SOAP client is tested against an `httptest` server; no real AzerothCore or
  SOAP server is required.
- For end-to-end development there is a fake worldserver at
  `internal/fake/azerothcore` (`cmd/fakeazerothcore`, `task fake:ac`). It speaks
  the same SOAP contract, keeps account state in memory and logs every command.
  See [runbooks/fake-azerothcore.md](runbooks/fake-azerothcore.md).

## See also

- [security.md](security.md)
- [ADR 0006](ADR/0006-module-owned-azeroth-commands.md)
- [database.md](database.md)
