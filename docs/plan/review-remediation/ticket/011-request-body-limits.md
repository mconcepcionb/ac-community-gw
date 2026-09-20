# 011 — Request body size limits

**Phase:** 3 · **Gate:** G-standard, contributes to G3-hardening · **Depends on:** 001

## Goal

Actually enforce the configured request body limit and stop draining unbounded
request bodies.

## Findings addressed

- Medium: `httpapi.DecodeJSON` wraps the body in `io.LimitReader` but the
  deferred `io.Copy(io.Discard, r.Body)` reads the whole remaining body. A client
  can send a valid JSON prefix followed by a huge tail; the request is accepted
  and the handler goroutine blocks draining the tail until `ReadTimeout`.

## Context

`internal/core/httpapi/response.go`. `DecodeJSON` currently takes only
`*http.Request`, so it cannot write a `413`.

## Atomic change

Switch to `http.MaxBytesReader` and return `413 request_too_large` reliably.
Change the decoder signature to accept the `ResponseWriter` (or return a typed
error the caller maps to 413).

## Requirements

- Use `r.Body = http.MaxBytesReader(w, r.Body, maxRequestBody)` before decoding.
- On overflow, return `httpapi.ErrRequestTooLarge` (`413`) with the standard
  error envelope; do not attempt to read further.
- Drain at most a small bounded amount (`io.CopyN`) during keep-alive cleanup.
- Keep `DisallowUnknownFields` and the single-object check.
- Make `maxRequestBody` a configurable value (reuse the existing config if one
  exists; otherwise add it in ticket 012).
- Apply the same guard to any other JSON decoder entry points (e.g. OAuth
  callback form parsing) or document why a path is exempt.

## Tests

- Test: body larger than the limit returns `413`.
- Test: valid JSON prefix plus a large tail returns `413` and does not hang
  (guard with a test timeout).
- Test: unknown field still `400`; two JSON objects still `400`.
- Test: a body exactly at the limit succeeds.

## Acceptance criteria (gate G-standard, G3-hardening)

- `task check` green; a `httptest` request with a 100 MB tail returns promptly
  with `413`.

## Rollback

Revert; the DoS window returns.

## Out of scope

- Per-route limits.
- Multer/streaming uploads (none exist).
