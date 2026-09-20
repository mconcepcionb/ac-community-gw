-- name: ListActiveProducts :many
SELECT * FROM store_products WHERE active = true ORDER BY price_points, name;

-- name: GetProductBySKU :one
SELECT * FROM store_products WHERE sku = $1;

-- name: InsertProduct :one
INSERT INTO store_products (id, sku, name, description, price_points, money, active)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateProduct :one
UPDATE store_products
SET name = $2, description = $3, price_points = $4, money = $5, active = $6, updated_at = now()
WHERE sku = $1
RETURNING *;

-- name: SetProductActive :one
UPDATE store_products
SET active = $2, updated_at = now()
WHERE sku = $1
RETURNING *;

-- name: DeleteProductItems :exec
DELETE FROM store_product_items WHERE product_id = $1;

-- name: InsertProductItem :exec
INSERT INTO store_product_items (product_id, item_id, count) VALUES ($1, $2, $3);

-- name: ListProductItems :many
SELECT item_id, count FROM store_product_items WHERE product_id = $1 ORDER BY item_id;

-- name: EnsureWallet :exec
INSERT INTO store_wallets (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING;

-- name: GetWallet :one
SELECT balance FROM store_wallets WHERE user_id = $1;

-- name: DebitWallet :one
UPDATE store_wallets
SET balance = balance - $2, updated_at = now()
WHERE user_id = $1 AND balance >= $2
RETURNING balance;

-- name: CreditWallet :one
UPDATE store_wallets
SET balance = balance + $2, updated_at = now()
WHERE user_id = $1
RETURNING balance;

-- name: InsertWalletEntry :exec
INSERT INTO store_wallet_entries (user_id, delta, balance_after, reason, order_id, actor_id)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: InsertOrder :one
INSERT INTO store_orders (id, user_id, product_id, sku, price_points, character_name, account_id, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpdateOrderStatus :one
UPDATE store_orders
SET status = $2, command_output = $3, updated_at = now()
WHERE id = $1 AND status = 'pending'
RETURNING *;

-- name: GetOrder :one
SELECT * FROM store_orders WHERE id = $1;

-- name: SetOrderOutput :exec
UPDATE store_orders
SET command_output = $2, updated_at = now()
WHERE id = $1 AND status = 'pending';

-- name: ListPendingOrdersWithOutput :many
SELECT * FROM store_orders
WHERE status = 'pending' AND command_output <> ''
ORDER BY created_at ASC
LIMIT $1;

-- name: ListOrdersByUser :many
SELECT * FROM store_orders WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;
