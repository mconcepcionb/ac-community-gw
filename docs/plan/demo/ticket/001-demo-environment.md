# One-command demo environment

## Goal

A `task demo` entrypoint that prepares the data and runs the fake AzerothCore,
the gateway and the SPA together.

## Requirements

- `task demo:setup`: `docker compose up -d --wait postgres mariadb`, migrate,
  seed.
- `task demo`: `demo:setup`, start `cmd/fakeazerothcore` in the background, run
  `task dev` in the foreground, print the SPA and dashboard URLs, and stop the
  fake server on exit.
- Scripts in `sh` and `ps1`; `.demo/` gitignored for the fake-server logs.

## Acceptance criteria

- `task demo` opens a logged-in SPA at `http://localhost:5173/` and a live fake
  AzerothCore dashboard at `http://localhost:7878/`.
- Ctrl+C stops the background fake server.

## Dependencies

- 002 (seeds).
