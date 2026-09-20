# 019 — Upstream transport hardening (SOAP and Discord)

**Phase:** 4 · **Gate:** G-standard, contributes to G4-data · **Depends on:** 012

## Goal

Harden the outbound transports: require TLS for credentialed upstreams, never
let an injected HTTP client silently lose its timeout, and treat malformed SOAP
responses as errors.

## Findings addressed

- Medium: Discord `TokenURL`/`APIBaseURL`/`AuthorizeURL` and `SOAPURL` may be
  `http://` even in production; only the redirect URL is forced to HTTPS, so the
  Discord client secret and SOAP Basic-auth credentials can transit cleartext.
- Low: supplying `HTTPClient` overrides `Timeout`; a client without a timeout can
  hang forever.
- Low: a SOAP `200` without an `executeCommandResponse`/`Fault` element is treated
  as success (`("", nil)`).
- Low: non-200 SOAP bodies (up to 256 bytes) are embedded into error strings that
  get logged.
- Nit: the manual SOAP envelope declares unused `xsi`/`xsd` namespaces.

## Context

`internal/adapters/azerothsoap/client.go`, `internal/adapters/discord/client.go`,
and `config.validate`. The SOAP body is XML-escaped and the Discord client
already avoids logging tokens.

## Atomic change

Add production URL checks in config validation and fix the two client defects.

## Requirements

- When `ACGW_ENV=production`, require `https://` for `AuthorizeURL`, `TokenURL`,
  `APIBaseURL` and `SOAPURL`; allow `http://` only for loopback/development.
- When an `HTTPClient` is injected without a timeout, clone it and apply
  `cfg.Timeout` (or fail validation); document the precedence.
- In the SOAP client, return a typed `ErrProtocol` when the response element is
  absent.
- Do not include upstream body content in errors unless a debug flag is enabled;
  log the request id and status only.
- Drop unused namespaces from the envelope.
- Keep command XML-escaping and the response size cap.

## Tests

- Config test: `http://` upstream URLs fail validation in production and pass in
  development.
- SOAP test: a `200` envelope without a response element yields `ErrProtocol`.
- SOAP test: a non-200 body does not appear in the returned error string.
- Client test: injected `HTTPClient` still enforces the timeout.

## Acceptance criteria (gate G-standard, G4-data)

- `task check` green.
- No credentialed upstream can be configured over cleartext in production.
- No upstream body content can reach logs via error strings.

## Rollback

Revert; the cleartext/hang/protocol risks return. Config checks are the only
user-visible change.

## Out of scope

- Retry/backoff policy.
- TLS pinning.
