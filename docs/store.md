# Store

The `store` plugin owns the community store: a points wallet per
community user, a product catalog and orders. Points live in the gateway
database; rewards are delivered in-game through the `azeroth-character`
delivery capability.

## Data model

Gateway-owned PostgreSQL (migration `00005_store.sql`):

| Table | Purpose |
| --- | --- |
| `store_products` | catalog entry: `sku`, `name`, `price_points`, `money`, `active` |
| `store_product_items` | item stacks granted by a product (`item_id`, `count`) |
| `store_wallets` | one balance per community user |
| `store_wallet_entries` | append-only ledger of every balance change |
| `store_orders` | a purchase and its delivery outcome |

A product's reward is a list of items and/or an amount of money (copper). Item
entries are picked from the [item catalog](azerothcore-integration.md#item-catalog)
(`GET /api/v1/azeroth/items?filter=`).

## Endpoints

| Method | Path | Permission |
| --- | --- | --- |
| `GET` | `/api/v1/store/products` | `gw.store.catalog.read` |
| `GET` | `/api/v1/store/products/{sku}` | `gw.store.catalog.read` |
| `POST` | `/api/v1/store/products` | `gw.store.admin.products` |
| `PUT` | `/api/v1/store/products/{sku}` | `gw.store.admin.products` |
| `DELETE` | `/api/v1/store/products/{sku}` | `gw.store.admin.products` |
| `GET` | `/api/v1/store/wallet` | `gw.store.wallet.read` |
| `GET` | `/api/v1/store/orders` | `gw.store.orders.read` |
| `POST` | `/api/v1/store/orders` | `gw.store.purchase` |
| `POST` | `/api/v1/store/wallets/grant` | `gw.store.admin.wallets` |

`POST /api/v1/store/orders` body `{"sku": "...", "character": "..."}`.
`POST /api/v1/store/wallets/grant` body `{"user_id": "..." | "discord_id": "...",
"points": 100, "reason": "..."}`.

## Curated items and rendering

The store sells a **curated set** of item entries taken from the item catalog.
`POST /api/v1/store/products` / `PUT .../{sku}` accept a bundle of items:

```json
{"sku": "starter-pack", "price_points": 300, "items": [{"item_id": 4496, "count": 1}, {"item_id": 6948, "count": 1}]}
```

- Each `item_id` is validated against the item catalog; unknown entries return
  `422 unknown_item`.
- A single-item product defaults its `sku` (`item-<entry>`), `name` and
  `description` from the item when omitted; bundles require an explicit `sku`.
- Responses embed a full **`render`** of each item (name, quality name/colour,
  class/subclass, inventory type, stats, damage with DPS, resistances, spells,
  binding, prices, description) so the storefront and the admin UI render an item
  with the same shape, without extra calls. Item icons require client DBC data
  and are not included; the `display_id` is exposed instead.
- `DELETE` soft-deletes (sets `active=false`), keeping order history intact.

## Purchase flow

1. Resolve the buyer's linked account (`azeroth.account.directory`). Without a
   link the request fails with `409 account_not_linked`.
2. **Transaction 1**: debit points with `UPDATE ... WHERE balance >= price`
   (atomic; no funds -> `402 insufficient_funds`), insert a `pending` order and a
   wallet ledger entry.
3. Deliver through `azeroth.character.delivery`, which verifies the character
   belongs to the account before running `.send items` / `.send money`.
4. **Transaction 2**: mark the order `delivered`, or on failure mark it `failed`
   and **refund** the points (plus a ledger entry). Points are never lost when
   delivery fails.

The order state machine is `pending -> delivered | failed` and is enforced in
SQL: status transitions only apply from `pending`, so completion is idempotent
and a refund can never be applied twice or to a delivered order. If delivery
succeeds but marking the order delivered fails, the delivery output is persisted
on the still-pending order and a periodic reconciliation task completes it; it is
never refunded.

Outcomes are written to the audit log. Administrative and account-management
commands are also audited.

## Deploying

```bash
task db:migrate
psql "$ACGW_DATABASE_URL" -f scripts/seed_demo_store.sql   # example catalog
```

Grant `gw.store.*` permissions to a role (the demo seed
`scripts/seed_demo_admin.sql` grants them to `ac-core.admin`).

## Out of scope (v1)

Multiple currencies, bundles/categories, coupons, idempotency keys, admin catalog
CRUD and event-driven point awards (login/playtime). See the plan in
`docs/plan/` when these are added.

## See also

- [azerothcore-integration.md](azerothcore-integration.md)
- [permissions.md](permissions.md)
