# Security

## Trust boundaries

```
Internet
   |
 HTTPS
   |
ac-community-gw
   |
 private network
   |
AzerothCore SOAP
```

- AzerothCore SOAP is never exposed to the Internet. It must live on the
  loopback interface, a private Docker network, a private network or a
  VPN/tunnel.
- The gateway is the only public interface to these capabilities.
- There is never an endpoint like `POST /api/v1/execute-command`, and no
  generic command endpoint of any kind. The API exposes only explicit,
  authorized domain operations.

## SOAP

- HTTP Basic Auth credentials are held only in configuration.
- The SOAP adapter is transport-only; it never exposes command names or syntax.
- Timeouts are enforced on every call.
- Dynamic values embedded in a command are validated with the shared
  `azerothcore.Safe*` helpers before the command string is built, and the
  executor refuses any command containing a line break (`ValidateCommand`).

## Sessions and cookies

- Opaque, high-entropy server-side session IDs.
- `HttpOnly`, `Secure`, `SameSite`.
- No sensitive data in the cookie.
- Authorization roles are refreshed on every session resolve (with a short
  cache), so a role change takes effect without waiting for the session TTL.

## Secrets handling

Never logged, never stored in events, never written to the audit log:

- passwords
- Discord access/refresh tokens
- session IDs
- SOAP credentials
- database credentials
- any secret

Configuration is read from environment variables. `.env` files are ignored by
git; only `.env.example` is committed.

Secrets are moving to **SOPS + age**: encrypted `secrets/*.sops.env` files are
committed and decrypted on the machine that needs them, with the private age
identities held outside the repository. Only `*.sops.env` (encrypted) and
`*.env.example` (templates) may be committed. The rollout, recipients and the
plaintext leak guard are specified in
[plan/secrets/README.md](plan/secrets/README.md).

## Audit

Administrative and sensitive actions can record:

```
timestamp, actor, action, permission, target type, target id, result,
request id, metadata
```

The audit boundary is `audit.Recorder`; a PostgreSQL-backed recorder is the
planned production implementation.

## HTTP hardening

- panic recovery returns a JSON 500
- every request has a request id (`X-Request-Id`), propagated in errors and logs;
  a client-supplied id is only accepted when it is short and log-safe
- consistent JSON error envelope
- request bodies are size-limited (`http.MaxBytesReader`, `413`) and reject
  unknown fields
- authentication endpoints are rate-limited per client IP (`429` + `Retry-After`);
  `X-Forwarded-For` is only trusted for proxies listed in
  `ACGW_TRUSTED_PROXIES` (empty by default)
- the optional `/metrics` endpoint is disabled by default and **requires**
  `ACGW_METRICS_TOKEN` when `ACGW_ENV=production`; in development it may run
  without a token and logs a warning
- `/readyz` reports only coarse per-check statuses; detailed errors are logged
  server-side
- startup refuses unsafe production configuration (insecure session cookie,
  non-HTTPS Discord redirect URL, cleartext credentialed upstream URLs, metrics
  without a token)
- the edge (Caddy) sets a restrictive `Content-Security-Policy`,
  `X-Frame-Options: DENY`, HSTS and `nosniff`, and does not proxy `/metrics`
- the development database containers bind to `127.0.0.1` and require explicit
  credentials from the environment (no defaults)

## See also

- [azerothcore-integration.md](azerothcore-integration.md)
- [authentication.md](authentication.md)
- [permissions.md](permissions.md)
