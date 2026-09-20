# ac-community-gw

A REST gateway between community applications and AzerothCore.

`ac-community-gw` provides a modern, authenticated and authorized HTTP API that
hides the internals of AzerothCore: its administrative SOAP interface and the
CLI commands it transports. External consumers never need to know about SOAP,
CLI syntax, AzerothCore schemas or administrative credentials.

The project is a **modular monolith** built as `core` + static plugins compiled
into a single binary.

## Requirements

- Go 1.27+
- [Task](https://taskfile.dev) as the development interface
- PostgreSQL 16+ (or Docker)
- [sqlc](https://sqlc.dev) for code generation
- [Goose](https://github.com/pressly/goose) for migrations
- Docker + Docker Compose (optional)
- Node 22+ and pnpm (optional, for the SPA under `web/`)

## Quickstart

```bash
cp .env.example .env
task db:up          # start PostgreSQL
task db:migrate     # apply migrations with Goose
task run            # start the gateway on :8080
```

Check it:

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz
```

Or run the whole development stack:

```bash
task docker:up
```

### SPA

The SPA lives under `web/` and is a separate artifact from the gateway. Install
once, then run the API and Vite (HMR) together:

```bash
task web:install
task dev            # API on :8080 + Vite on :5173 with /api proxied
```

In production Caddy serves the built SPA and reverse-proxies the API on the
same origin (`web/Caddyfile`; `task docker:up` runs the whole stack). The
gateway never serves the SPA. See [docs/frontend.md](docs/frontend.md) and
[ADR 0012](docs/ADR/0012-decoupled-spa-serving.md).

## Taskfile

| Task | Description |
| --- | --- |
| `task run` | Run the gateway locally |
| `task dev` | Run the API and the Vite dev server (HMR) together |
| `task build` | Build `bin/ac-community-gw` (Go only) |
| `task test` | Run unit tests |
| `task test:race` | Run tests with the race detector |
| `task test:integration` | Run integration tests (PostgreSQL + MySQL) |
| `task coverage` | Run tests with coverage |
| `task fmt` / `task fmt:check` | Format / verify formatting |
| `task vet` | Run `go vet` |
| `task check` | Format check + vet + tests + OpenAPI drift + SPA checks |
| `task codegen` / `task codegen:sqlc` | Regenerate sqlc code |
| `task openapi` / `task openapi:check` | Regenerate / verify the spec and TS client |
| `task web:dev` / `task web:build` | SPA dev server / build into `web/dist` |
| `task web:test` / `task web:lint` / `task web:check` | SPA tests / lint / all |
| `task web:image` | Build the SPA reverse-proxy (Caddy) image |
| `task db:up` / `task db:down` | Start / stop PostgreSQL |
| `task db:migrate` / `task db:rollback` / `task db:status` | Goose migrations |
| `task docker:build` | Build the production API image |
| `task docker:up` / `task docker:down` | Development stack (proxy + API + DB) |

## Repository layout

```
cmd/server            entrypoint and wiring
internal/core         transversal infrastructure (no domain knowledge)
internal/plugins      compiled-in domain modules
internal/adapters     external integrations (PostgreSQL, AzerothCore SOAP)
api                   generated OpenAPI spec (source for the TypeScript client)
web                   SPA (React + TypeScript + Vite) + Caddy proxy
migrations            single Goose migration sequence
docs                  stable docs, ADRs, plans, runbooks
```

The SPA is a first-class consumer of the REST API: the contract is generated
from Go annotations. It is built and served separately from the gateway, by
Caddy, on the same origin. See [docs/frontend.md](docs/frontend.md).

## Architecture in one minute

- `core` MUST NOT import plugins.
- Plugins MAY import core contracts.
- A plugin MUST NOT import another plugin's implementation.
- Cross-plugin interaction happens through core registries:
  - **Query** → service/capability registry
  - **Command** → application command registry
  - **Event** → synchronous in-process event bus
  - **Permission** → permission registry + RBAC
- The SOAP adapter transports commands; it does not own command syntax.
  Each plugin owns the AzerothCore syntax for the capabilities it implements.

See `docs/architecture.md` for details.

## Configuration

All configuration is read from environment variables. See `.env.example`.
Secrets are never committed.

## Tests

```bash
task test
task check
```

Unit tests do not require Discord, AzerothCore, SOAP or PostgreSQL. The test
suite also enforces the import boundaries of the modular monolith.

## Migrations

```bash
task db:migrate
task db:status
task db:rollback
```

The server can also apply migrations at startup when `ACGW_DB_AUTO_MIGRATE=true`.

## sqlc

Queries are owned per module under `internal/plugins/<module>/repository/`.
Regenerate with:

```bash
task codegen:sqlc
```

## Documentation

- Stable behaviour: `docs/*.md`
- Decisions: `docs/ADR/`
- Plans and tickets: `docs/plan/`
- Operations: `docs/runbooks/`
