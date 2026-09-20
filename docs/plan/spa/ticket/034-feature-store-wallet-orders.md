# Feature: store wallet and orders

## Goal

Show the wallet, place orders and grant wallet balance.

## Context

Consumes `GET /api/v1/store/wallet`, `GET/POST /api/v1/store/orders`
(ticket 017) and `POST /api/v1/store/wallets/grant` (ticket 018).

## Requirements

- Route `/store/wallet`: balance + order history table.
- Order creation flow (product + quantity/recipient) with confirmation and
  clear success/failure feedback.
- Admin route/action `/store/wallets/grant` (permission-gated) with
  confirmation.
- Invalidate wallet and orders after mutations.

## Acceptance criteria

- Wallet and orders render from mocked data; order creation and grant call the
  right operations.
- Grant is hidden without permission and always confirmed.
- `task web:check` green.

## Dependencies

- 017, 018, 020, 022, 023.
