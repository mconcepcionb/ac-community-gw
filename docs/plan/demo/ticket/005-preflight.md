# Demo preflight

## Goal

Fail fast with a clear message when the environment is not demo-ready.

## Requirements

- Before starting, verify `.env` exists and the required Discord variables
  (`CLIENT_ID`, `CLIENT_SECRET`, `REDIRECT_URL`, `GUILD_ID`) are set.
- Point at `docs/runbooks/discord-oauth-setup.md` on failure.
- Never print secret values.

## Acceptance criteria

- A missing configuration stops `task demo` with actionable guidance.

## Dependencies

- 001.
