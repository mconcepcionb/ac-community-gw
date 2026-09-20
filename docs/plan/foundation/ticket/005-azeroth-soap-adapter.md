# AzerothCore SOAP adapter

## Goal

Provide a replaceable, transport-only SOAP client implementing the
`CommandExecutor` contract.

## Context

AzerothCore exposes a small SOAP `executeCommand` interface. The adapter must
hide SOAP details from plugins and never own command syntax. See ADR 0006.

## Requirements

- Contract: `type CommandExecutor interface { Execute(ctx, command) (string, error) }`.
- `net/http` only, no SOAP library.
- Build and parse SOAP XML by hand.
- HTTP Basic Auth.
- Configurable timeout.
- Translate HTTP status, SOAP faults and protocol errors into typed errors.
- The adapter must not be imported by plugins.

## Acceptance criteria

- Successful command returns the result text.
- SOAP fault returns an `ErrSOAPFault` error containing the fault detail.
- Non-200 without a fault returns `ErrHTTPStatus`.
- Invalid XML returns `ErrProtocol`.
- Timeouts produce an error.
- Special characters in commands are XML-escaped.

## Implementation notes

`cmd/server` injects the client; when no SOAP URL is configured it injects
`azerothcore.Unavailable`.

## Tests

`net/http/httptest` fake SOAP server covering success, fault, HTTP error,
protocol error, timeout and escaping. No real AzerothCore required.
