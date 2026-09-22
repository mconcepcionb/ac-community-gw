#!/usr/bin/env sh
# Seed the demo data. Idempotent; safe to re-run.
set -eu

cd "$(git rev-parse --show-toplevel)"

psql_file() {
    # $1 local sql file, remaining args are psql -v options.
    file="$1"
    shift
    docker compose cp "$file" postgres:/tmp/acgw-demo.sql >/dev/null
    docker compose exec -T postgres psql -U acgw -d acgw -v ON_ERROR_STOP=1 "$@" -f /tmp/acgw-demo.sql
}

echo "demo: seeding the store catalog"
psql_file scripts/seed_demo_store.sql

role_id="${ACGW_DEMO_DISCORD_ROLE_ID:-}"
if [ -n "$role_id" ]; then
    echo "demo: mapping Discord role $role_id to the internal role ac-core.admin"
    psql_file scripts/seed_demo_admin.sql -v discord_role_id="$role_id" -v internal_role=ac-core.admin
else
    echo "demo: ACGW_DEMO_DISCORD_ROLE_ID is not set; skipped the admin role mapping"
    echo "demo: set it in .env to give the demo account staff permissions"
fi
