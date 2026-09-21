# Development

## Requirements

- Go 1.27+
- Task
- sqlc
- Goose
- golangci-lint 2.x (see `.golangci.yml`)
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
[ADR 0015](ADR/0015-sops-age-secrets.md) and
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
task coverage:check    # fail when measured coverage drops below coverage.floor
task coverage:integration  # unit + integration profiles
task fmt
task fmt:check
task vet
task lint              # golangci-lint (see .golangci.yml)
task check          # fmt:check + vet + test + test:race + openapi:check + web:check
task codegen:sqlc
task openapi        # regenerate api/swagger.yaml and the TS client
task openapi:check
task web:dev
task web:build
task web:test
task web:coverage    # vitest --coverage with thresholds (web/vite.config.ts)
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

### Coverage

`task coverage` writes `coverage.txt` and prints a per-package table produced by
`cmd/covercheck`, which measures only the **measured set**: `./internal/...`
minus the prefixes listed in `coverage.ignore` (generated code, the
test-only architecture package, the fake AzerothCore double and `cmd/`). Those
excluded packages are still shown, but they are not counted in the total.

`task coverage:check` (also run in CI) fails when the total falls below
`coverage.floor` (currently **61.0**). The floor only ratchets up: raise it in
the same change that adds coverage. `task coverage:integration` merges the
`-tags integration` profile into the total (currently **76.3%**).

The SPA gate is separate: `task web:coverage` enforces 60% statements/lines over
`web/src` (currently **68.4%** statements) and excludes the generated client and
`routeTree.gen.ts`. CI runs `task coverage:check` in the Go job,
`pnpm test:coverage` in the web job, and `task test:integration` in the
integration job.

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
- `task web:coverage` runs the Vitest coverage gate (60% statements/lines over
  `web/src`, excluding the generated client and `routeTree.gen.ts`). It runs
  test files serially so the threshold is stable on slow machines; CI runs the
  same command.

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
