# ADR 0002: net/http only

## Status

Accepted

## Context

Go's standard library `net/http`, with `http.ServeMux` method patterns
(`"GET /path"`), is sufficient for the API surface we need: routing, middleware,
JSON handling, timeouts and graceful shutdown.

## Decision

Use exclusively `net/http` and `http.ServeMux`. Do not introduce Gin, Echo,
Fiber, Chi, Gorilla, FastHTTP or equivalents.

## Consequences

- No external HTTP dependency and no framework upgrade risk.
- Middleware and helpers are written by us and stay small.
- Some conveniences (route groups, path parameters beyond the stdlib patterns)
  are not available; the stdlib patterns cover our needs.
