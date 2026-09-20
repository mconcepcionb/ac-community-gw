# ADR 0004: Server-side sessions

## Status

Accepted

## Context

The application is a web-facing gateway. Authentication is delegated to Discord
via OAuth2. We need a session mechanism that is easy to revoke and safe in the
browser.

## Decision

Use **server-side sessions** stored in PostgreSQL, referenced by a random
opaque token in an `HttpOnly`, `Secure`, `SameSite` cookie. Do not introduce JWT
for now.

## Consequences

- Sessions can be revoked immediately server-side.
- No sensitive data is stored in the cookie.
- Each authenticated request performs a session lookup.
- A PostgreSQL-backed store is planned; this iteration ships an in-memory store
  plus the persistence contract.
