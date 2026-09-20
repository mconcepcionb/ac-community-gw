# SPA fallback

## Goal

Serve the built SPA from the gateway, falling back to `index.html` for
client-side routes while preserving the JSON API envelope.

## Context

`internal/core/httpapi/server.go` mounts `http.FileServer` at `/` when
`ACGW_WEB_DIR` is set. A file server returns 404 for client routes such as
`/store/products`, which the SPA router must handle.

## Requirements

- Replace the bare `http.FileServer` mount with a handler that:
  - serves existing static files from the directory;
  - returns `index.html` for extension-less paths outside `/api/`;
  - keeps returning the JSON error envelope for `/api/*` and for missing static
    assets (paths with an extension);
  - keeps the "web disabled" behaviour: no directory -> JSON 404 at `/`.
- No change to `/healthz`, `/readyz` or `/metrics` routing.

## Acceptance criteria

- Existing tests pass unchanged: `TestWebServesIndex`,
  `TestWebKeepsJSONErrorsForUnknownAPIPaths`, `TestWebDisabledReturnsJSONNotFound`.
- New: `GET /some/client/route` returns the index document with 200.
- New: `GET /assets/missing.js` returns the JSON 404 envelope, not the index.
- New: `GET /api/v1/unknown` still returns the JSON 404 envelope.

## Implementation notes

- Keep the implementation in `httpapi` (core, no domain knowledge); do not
  introduce a dependency on the SPA source.
- Guard against path traversal (`http.Dir` already does) and only treat
  extension-less paths as routes.

## Tests

- Table test over the cases above using a temp dir with `index.html` and one
  asset file.

## Dependencies

- None (independent of 001/002). Unblocks 019 and 035.
