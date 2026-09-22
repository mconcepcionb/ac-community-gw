#!/usr/bin/env sh
# Grant demo points to the user identified by ACGW_DEMO_DISCORD_ID.
set -eu

cd "$(git rev-parse --show-toplevel)"

discord_id="${ACGW_DEMO_DISCORD_ID:-}"
if [ -z "$discord_id" ]; then
    echo "demo: set ACGW_DEMO_DISCORD_ID (your Discord user id) to grant points" >&2
    exit 1
fi
points="${ACGW_DEMO_POINTS:-1000}"

docker compose cp scripts/seed_demo_grant.sql postgres:/tmp/acgw-demo-grant.sql >/dev/null
docker compose exec -T postgres psql -U acgw -d acgw -v ON_ERROR_STOP=1 \
    -v discord_id="$discord_id" -v points="$points" -f /tmp/acgw-demo-grant.sql
echo "demo: granted $points points to Discord user $discord_id"
