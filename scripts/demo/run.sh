#!/usr/bin/env sh
# Run the full local demo: prepare the data, start the fake AzerothCore in the
# background, then run the gateway and the SPA (task dev) in the foreground.
set -eu

cd "$(git rev-parse --show-toplevel)"

if [ ! -f .env ]; then
    echo "demo: missing .env" >&2
    echo "demo: configure Discord first (see docs/runbooks/discord-oauth-setup.md)" >&2
    exit 1
fi

missing=""
for key in ACGW_DISCORD_CLIENT_ID ACGW_DISCORD_CLIENT_SECRET ACGW_DISCORD_REDIRECT_URL ACGW_DISCORD_GUILD_ID; do
    if ! grep -q "^${key}=..*" .env 2>/dev/null; then
        missing="$missing $key"
    fi
done
if [ -n "$missing" ]; then
    echo "demo: missing required Discord config in .env:$missing" >&2
    echo "demo: see docs/runbooks/discord-oauth-setup.md" >&2
    exit 1
fi

task demo:setup

mkdir -p .demo
echo "demo: starting the fake AzerothCore on http://localhost:7878/"
go run ./cmd/fakeazerothcore -addr 127.0.0.1:7878 -user acgw -pass acgw > .demo/fake-ac.log 2>&1 &
fake_pid=$!
cleanup() { kill "$fake_pid" 2>/dev/null || true; }
trap cleanup EXIT INT TERM

sleep 2
cat <<'BANNER'

===================================================================
 demo addresses
   SPA (login with Discord): http://localhost:5173/
   fake AzerothCore dash:    http://localhost:7878/
   walkthrough:              docs/demo.md
===================================================================

BANNER
task dev
