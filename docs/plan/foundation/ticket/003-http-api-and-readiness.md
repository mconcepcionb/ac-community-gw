# HTTP API and readiness

## Goal

Provide the HTTP server, middleware, consistent JSON errors and the operational
endpoints.

## Context

Consumers need a stable JSON API with request correlation, and operators need
health and readiness probes.

## Requirements

- `net/http` and `http.ServeMux` only, method patterns (`"GET /path"`).
- JSON responses, request decoding with size limit and unknown-field rejection.
- Consistent error envelope mapping to 400/401/403/404/409/422/500/502.
- Middleware: request id, panic recovery, structured access log, authentication,
  authorization.
- `GET /healthz` and `GET /readyz` outside `/api/v1`.
- `readyz` checks PostgreSQL and reports per-check status.

## Acceptance criteria

- `/healthz` returns 200 while the process is alive.
- `/readyz` returns 503 when a readiness check fails.
- Anonymous access to protected routes returns 401; missing permission returns
  403.
- Unknown routes return the JSON error envelope with a request id.

## Implementation notes

Readiness checks implement `persistence.Check`. Auth middleware resolves the
server-side session and stores the principal in the context.

## Tests

- health and readiness (ready/unavailable)
- require auth (401)
- require permission (403 and 200)
- unknown route JSON error with request id
