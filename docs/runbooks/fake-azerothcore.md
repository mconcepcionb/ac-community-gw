# Runbook: fake AzerothCore

## Purpose

Run a development double of the AzerothCore SOAP interface so the gateway (and
its `azeroth-account`, `azeroth-admin` and `azeroth-info` plugins) can be
exercised without a real worldserver.

The double speaks the same wire protocol as the real server, keeps the provided
account state in memory and logs every command. It never touches MySQL.

## Preconditions

- Go toolchain (or the built binary).
- Nothing listening on the chosen port (default `7878`).

## Procedure

Start it:

```bash
task fake:ac
# or, with a seed:
go run ./cmd/fakeazerothcore -state ./cmd/fakeazerothcore/seed.example.json
```

Flags:

| Flag | Default | Description |
| --- | --- | --- |
| `-addr` | `:7878` | listen address |
| `-user` | `acgw` | HTTP Basic Auth username (empty disables auth) |
| `-pass` | `acgw` | HTTP Basic Auth password |
| `-state` | empty | JSON file with initial accounts |
| `-journal` | `200` | number of commands kept in the in-memory journal |
| `-login-db-dsn` | `$ACGW_AZEROTH_LOGIN_DB_DSN` | optional AzerothCore login DB to mirror accounts into |

Point the gateway at it in `.env`:

```dotenv
ACGW_AZEROTH_SOAP_URL=http://localhost:7878/
ACGW_AZEROTH_SOAP_USERNAME=acgw
ACGW_AZEROTH_SOAP_PASSWORD=acgw
```

## Endpoints

| Method | Path | Purpose |
| --- | --- | --- |
| `POST` | `/` | SOAP `executeCommand` (same contract as the real server) |
| `GET` | `/` | live HTML dashboard (command journal + state) |
| `GET` | `/healthz` | liveness |
| `GET` | `/state` | current account state as JSON |
| `GET` | `/commands?limit=N` | command journal, newest first (default 50) |
| `GET` | `/commands/stream` | command journal as Server-Sent Events |
| `POST` | `/reset` | restore the seed accounts and clear the journal |

## Live view

Open `http://localhost:7878/` next to the gateway to watch the internal side of
each call: the dashboard streams every command the double receives, with its
result and current account state. Use `POST /reset` (or the dashboard button)
between tests to get a clean slate.

The SPA (`task dev`) also embeds this journal in its **Logs** section, side by
side with the gateway calls it makes. Because the SPA runs on a different
origin (Vite on :5173), the observability endpoints send permissive CORS
headers — another reason to keep the double on localhost.

### Request correlation

The gateway's SOAP adapter forwards the request id it received (or generated)
as the `X-Request-Id` header. The double records it, so the dashboard and
`GET /commands` show the same id that appears in the gateway's logs — this is
what lets you line up an API call with the internal AzerothCore command:

```
gateway log:  request_id=abc123 POST /api/v1/... -> .account create ...
fake journal: request_id=abc123 .account create ... -> Account created: ...
```


## Supported commands

Derived from AzerothCore `cs_account.cpp`, `cs_ban.cpp` and `cs_server.cpp`;
response texts come from the `acore_string` table.

```
.server info
.account create <user> <pass> [email]
.account set password <user> <pass> <pass>
.account set email <user> <email> [emailConfirmation]
.account set gmlevel <user> <level> [realm]
.ban account <user> <duration> <reason>
.unban account <user>
```

`<duration>` accepts `Nd`, `Nh`, `Nm`, `Ns` (e.g. `1d`, `2h30m`); `0` or an
unparsable value bans permanently.

## Validation

```bash
# health
curl -s http://localhost:7878/healthz

# a command with the same shape the gateway sends
curl -s -u acgw:acgw -X POST http://localhost:7878/ \
  -H 'Content-Type: text/xml; charset=utf-8' \
  -H 'SOAPAction: "urn:AC#executeCommand"' \
  --data '<?xml version="1.0"?><soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"><soap:Body><executeCommand xmlns="urn:AC"><commandString>.account create Test pass test@example.com</commandString></executeCommand></soap:Body></soap:Envelope>'

# inspect state
curl -s http://localhost:7878/state
```

The response is a SOAP envelope whose `<result>` is the command output:

```
Account created: Test
```

## Testing through the gateway

With the gateway running against this double, `GET /api/v1/azeroth/info/status`
returns the fake `.server info` output (it requires the
`azeroth.info.public.read` permission, so map a Discord role to it first).

Create accounts through the command registry from an integration test, or
invoke `.server info` via the status route.

## Login database mirror

When `-login-db-dsn` (or `ACGW_AZEROTH_LOGIN_DB_DSN`) is set, account commands
are mirrored into an AzerothCore login database so the gateway's read adapter
(`GET /api/v1/azeroth/accounts`) sees the same accounts the double serves over
SOAP. The compose MariaDB fixture reproduces the official login and character
base schemas (`data/sql/base/db_auth`, `db_characters`):

```bash
docker compose up -d mariadb
task fake:ac   # picks up ACGW_AZEROTH_LOGIN_DB_DSN from .env
```

`POST /reset` resets the in-memory state and journal only; it does not clear the
database.

## Limitations

- In-memory only: state resets on restart unless seeded with `-state`.
- The journal is a bounded ring buffer (default 200 entries).
- The observability endpoints (`/`, `/state`, `/commands`, `/commands/stream`,
  `/reset`) are unauthenticated, like a local debug tool; bind to localhost or a
  private network and never expose the double publicly.
- No characters, no realm/DBC data; `.server info` reports fixed values.
- `.account set email` accepts a single address (the gateway sends one) where
  the real handler requires address + confirmation.
- Security-level checks of the real server are not modelled; any authenticated
  caller may change any account.
- No rate limiting, no TLS: bind to localhost or a private network only.

## See also

- [azerothcore-integration.md](../azerothcore-integration.md)
- [deploy.md](deploy.md)
