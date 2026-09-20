# ADR 0007: Taskfile as development interface

## Status

Accepted

## Context

Developers need a single, discoverable interface for running, building, testing,
generating code, managing the database and using Docker. Makefiles are not
comfortable on Windows, one of our development platforms.

## Decision

Use **Task** (`Taskfile.yml`) as the standard development interface. Do not add
a Makefile. Prefer `task ...` over invoking Goose, sqlc or Docker directly when
an equivalent task exists.

## Consequences

- One documented entry point (`task --list`).
- Tasks are cross-platform.
- Runbooks reference tasks rather than raw tool invocations.
- Developers must install Task.
