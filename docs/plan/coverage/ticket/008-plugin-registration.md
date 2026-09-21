# Plugin registration

**Milestone:** B · **Gate:** G-standard · **Depends on:** 001, 003

## Goal

Verify that every plugin's `Register` wires its routes, permissions, commands,
services and events without collisions, and that a duplicate is rejected.

## Context

No test builds `plugins.Registry` or calls `Manager.RegisterAll`
(`internal/core/plugins/registry.go:32`). Route and permission wiring is
therefore unverified for every plugin, including the `permissionDefs` in each
`permissions.go`. A typo in a route or a duplicated permission name would ship
undetected.

## Requirements

- A test that constructs a `plugins.Manager`, adds every plugin with its minimal
  fake `Config` (reusing `internal/testsupport` and per-package fakes), calls
  `RegisterAll`, and asserts:
  - every plugin registers a unique name (`ErrDuplicateName` otherwise) and a
    non-empty name (`ErrEmptyName` otherwise);
  - the expected route set is present on the aggregate `Mux`;
  - every `permissionDefs` entry is registered and namespaced `gw.` or
    `<game>.` per ADR 0014;
  - no plugin registers the same permission or route twice.
- A negative test that adding two plugins with the same name fails.

## Tests

- The convergence test above; it is the single place that exercises `Register`
  end to end for all plugins.

## Acceptance criteria

- The test fails if a plugin omits a route or declares a mis-namespaced
  permission.
- `task check` green.

## Out of scope

- Deep per-route behavior (C4-C7) and repository wiring (C9).

## Notes

Keep the fakes minimal; if a plugin's `Register` needs a non-trivial config,
prefer adding a constructor-friendly fake in `internal/testsupport` over
importing the plugin into another plugin's test.
