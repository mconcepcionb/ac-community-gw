#!/usr/bin/env sh
# Reset the demo between runs: clear the fake AzerothCore state and re-seed.
set -eu

cd "$(git rev-parse --show-toplevel)"

if curl -fsS -X POST http://localhost:7878/reset >/dev/null 2>&1; then
    echo "demo: reset the fake AzerothCore state"
else
    echo "demo: fake AzerothCore is not running (start it with task demo)"
fi

sh scripts/demo/seed.sh
