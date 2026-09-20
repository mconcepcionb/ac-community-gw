-- +goose Up

CREATE TABLE store_products (
    id           uuid PRIMARY KEY,
    sku          text NOT NULL UNIQUE,
    name         text NOT NULL,
    description  text NOT NULL DEFAULT '',
    price_points bigint NOT NULL CHECK (price_points >= 0),
    money        bigint NOT NULL DEFAULT 0 CHECK (money >= 0),
    active       boolean NOT NULL DEFAULT true,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE store_product_items (
    product_id uuid NOT NULL REFERENCES store_products (id) ON DELETE CASCADE,
    item_id    integer NOT NULL,
    count      integer NOT NULL CHECK (count > 0),
    PRIMARY KEY (product_id, item_id)
);

CREATE TABLE store_wallets (
    user_id    uuid PRIMARY KEY REFERENCES community_users (id) ON DELETE CASCADE,
    balance    bigint NOT NULL DEFAULT 0 CHECK (balance >= 0),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE store_wallet_entries (
    id            bigserial PRIMARY KEY,
    user_id       uuid NOT NULL REFERENCES community_users (id) ON DELETE CASCADE,
    delta         bigint NOT NULL,
    balance_after bigint NOT NULL,
    reason        text NOT NULL,
    order_id      uuid,
    actor_id      uuid,
    created_at    timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX store_wallet_entries_user_id_idx ON store_wallet_entries (user_id, created_at DESC);

CREATE TABLE store_orders (
    id             uuid PRIMARY KEY,
    user_id        uuid NOT NULL REFERENCES community_users (id) ON DELETE CASCADE,
    product_id     uuid NOT NULL REFERENCES store_products (id) ON DELETE RESTRICT,
    sku            text NOT NULL,
    price_points   bigint NOT NULL,
    character_name text NOT NULL,
    account_id     bigint NOT NULL,
    status         text NOT NULL,
    command_output text NOT NULL DEFAULT '',
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX store_orders_user_id_idx ON store_orders (user_id, created_at DESC);

-- +goose Down

DROP TABLE store_orders;
DROP TABLE store_wallet_entries;
DROP TABLE store_wallets;
DROP TABLE store_product_items;
DROP TABLE store_products;
