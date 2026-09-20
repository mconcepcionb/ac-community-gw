-- Example store catalog for local development. Idempotent: safe to re-run.
--
--   psql "$ACGW_DATABASE_URL" -f scripts/seed_demo_store.sql
--   # or: Get-Content scripts/seed_demo_store.sql -Raw | docker compose exec -T postgres psql -U acgw -d acgw

INSERT INTO store_products (id, sku, name, description, price_points, money, active) VALUES
    ('11111111-1111-1111-1111-111111111111', 'bag-16',
     'Traveler''s Backpack', 'A 16-slot bag delivered by mail.', 100, 0, true),
    ('22222222-2222-2222-2222-222222222222', 'gold-1000',
     '1000 Gold', '1000 gold delivered by mail.', 250, 10000000, true),
    ('33333333-3333-3333-3333-333333333333', 'starter-pack',
     'Starter Pack', 'A bag and some gold.', 300, 5000000, true)
ON CONFLICT (sku) DO UPDATE
SET name = EXCLUDED.name,
    description = EXCLUDED.description,
    price_points = EXCLUDED.price_points,
    money = EXCLUDED.money,
    active = EXCLUDED.active,
    updated_at = now();

-- Items are keyed on the product SKU so this stays idempotent even when a
-- product already exists under a different id.
INSERT INTO store_product_items (product_id, item_id, count)
SELECT p.id, v.item_id, v.count
FROM (VALUES
    ('bag-16', 4496, 1),
    ('starter-pack', 4496, 1)
) AS v(sku, item_id, count)
JOIN store_products p ON p.sku = v.sku
ON CONFLICT (product_id, item_id) DO UPDATE SET count = EXCLUDED.count;
