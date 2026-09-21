# Development

## Requirements

- Go 1.27+
- Task
- sqlc
- Goose
- Docker + Docker Compose (optional)
- PostgreSQL 16+ (or `task db:up`)
- Node 22+ and pnpm (only for the SPA under `web/`)

## Setup

```bash
cp .env.example .env
task db:up
task db:migrate
task run
```

`task run` loads `.env` through Task's `dotenv` support.

Secrets are managed with **SOPS + age** (see
[plan/secrets/README.md](plan/secrets/README.md)). Once that plan lands, the
plaintext `.env` is produced from the encrypted set with
`task secrets:decrypt ENV=development` instead of copied from `.env.example`;
until then, keep using the `cp` step above.

For the SPA, install its dependencies once:

```bash
task web:install
```

## Standard tasks

```bash
task run
task dev            # API + Vite dev server (HMR) together
task build
task test
task test:race
task test:integration  # needs ACGW_DATABASE_URL and the MySQL fixtures
task coverage
task fmt
task fmt:check
task vet
task check          # fmt:check + vet + test + test:race + openapi:check + web:check
task codegen:sqlc
task openapi        # regenerate api/swagger.yaml and the TS client
task openapi:check
task web:dev
task web:build
task web:test
task web:lint
task web:check
task db:up
task db:down
task db:migrate
task db:rollback
task db:status
task web:image      # build the SPA reverse-proxy image
task docker:build
task docker:up
task docker:down
```

Use `task ...` rather than invoking Goose, sqlc, pnpm or Docker directly when a
task exists.

## Testing

```bash
task test
```

- Unit tests never require Discord, AzerothCore, SOAP or PostgreSQL.
- HTTP handlers are tested with `httptest`.
- The SOAP transport is tested with a fake SOAP server.
- The architectural import boundaries are enforced by tests in
  `internal/architecture`, including a `gofmt` check.

## Code generation

sqlc output is committed. After editing `queries.sql` or a migration:

```bash
task codegen:sqlc
task check
```

The OpenAPI spec and the TypeScript client are generated from the Go
annotations. After changing a handler annotation or a DTO:

```bash
task openapi        # api/swagger.yaml + web/src/api/generated
task check
```

## Frontend

The SPA lives under `web/` (React + TypeScript + Vite). See
[frontend.md](frontend.md) for the structure, toolchain, auth flow and testing.

- `task dev` runs the gateway and the Vite dev server together; the SPA is
  served by Vite on `:5173` with `/api` proxied to the gateway on `:8080`.
- `task web:build` builds the SPA into `web/dist`. Caddy serves it in front of
  the gateway (`web/Caddyfile`, `web/Dockerfile`); the gateway itself never
  serves the SPA.

## Conventions

- Prefer the standard library.
- Small, typed interfaces.
- No HTTP framework, no ORM, no logging framework, no DI framework.
- Explicit SQL.
- Plugins own their routes, permissions, commands, events, services and SQL.
- Do not add speculative abstractions.

## Code review checklist

- [ ] `task check` passes
- [ ] no new dependency without justification
- [ ] no cross-plugin imports
- [ ] no secrets in code, logs, events or audit entries
- [ ] documentation updated when behaviour changes
