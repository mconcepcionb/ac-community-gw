# Self-service account creation and linking

## Goal

Let a signed-in user with no game account create one and link it in a single
flow.

## Context

Linking is admin-only today. This ticket covers use case P2
([../../../use-cases.md](../../../use-cases.md)), guarded by Discord guild
membership and rate limits. It is a **Now** horizon ticket, so players can
complete onboarding before the Next-horizon features build on it.

## Requirements

- Backend: a create-and-link endpoint for the authenticated user.
  - Requires Discord guild membership.
  - Enforces one game account per community user.
  - Rate-limited per user and per IP.
  - Audited.
- Portal `/onboarding`: offer "create a new account" or "I already have one"
  (012).
- Creation takes a self-chosen username and password valid under AzerothCore
  rules; no generated credentials, nothing shown once.
- Inline, retryable errors for taken username, weak password and world
  unavailable.
- Annotate the endpoints and regenerate the OpenAPI spec and TS client.

## Acceptance criteria

- An eligible user creates and links in one flow and lands on their characters.
- An ineligible user (not in the guild) sees an explicit refusal.
- A second create or claim is refused with a clear explanation.
- The flow is rate-limited and audited.
- `task check`, `task openapi:check` and `task web:check` green.

## Implementation notes

- Reuse the `azeroth-account` create capability plus the link capability;
  ownership is implied by creation.
- Guild membership is already available: the Discord adapter requests the
  `guilds.members.read` scope and exposes the configured `ACGW_DISCORD_GUILD_ID`
  member lookup (`internal/adapters/discord/client.go`). Reuse it.
- Use typed application commands; never expose a raw AzerothCore command.

## Tests

- Go tests: eligibility, duplicate account, rate limit, audit.
- RTL + MSW for the wizard and error states.

## Dependencies

- 001, 002, 008-010. 012 shares the onboarding entry.
