# ADR 0010: Synchronous in-process event bus

## Status

Accepted

## Context

Plugins need to announce domain facts without knowing who consumes them. A
durable or distributed event infrastructure would add significant operational
complexity that is not justified yet.

## Decision

The event bus is:

- in-process
- in-memory
- synchronous
- non-durable

`Publish()` runs the registered handlers **within the same call**, in
registration order, and never starts goroutines automatically. Handler errors
are joined and returned to the publisher.

Kafka, NATS, RabbitMQ and Redis Streams are explicitly out of scope.

## Consequences

- Deterministic behaviour, visible errors, simple tests and simple shutdown.
- Delivery is best-effort for the lifetime of the process only.
- The bus must not be used for critical operations whose loss would leave the
  system inconsistent (for example charging points and then delivering an item).
- If durable guarantees are ever required, a transactional outbox should be
  considered explicitly and decided in a new ADR.
- Events must never carry secrets or sensitive data.
