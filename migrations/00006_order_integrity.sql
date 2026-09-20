-- +goose Up

ALTER TABLE store_orders
    ADD CONSTRAINT store_orders_status_check
    CHECK (status IN ('pending', 'delivered', 'failed'));

ALTER TABLE store_orders
    ADD CONSTRAINT store_orders_price_points_check
    CHECK (price_points >= 0);

-- +goose Down

ALTER TABLE store_orders DROP CONSTRAINT IF EXISTS store_orders_price_points_check;
ALTER TABLE store_orders DROP CONSTRAINT IF EXISTS store_orders_status_check;
