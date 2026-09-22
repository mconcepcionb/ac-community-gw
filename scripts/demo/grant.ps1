# Grant demo points to the user identified by ACGW_DEMO_DISCORD_ID.
$ErrorActionPreference = "Stop"

Set-Location (git rev-parse --show-toplevel)

if (-not $env:ACGW_DEMO_DISCORD_ID) {
    throw "demo: set ACGW_DEMO_DISCORD_ID (your Discord user id) to grant points"
}
$points = if ($env:ACGW_DEMO_POINTS) { $env:ACGW_DEMO_POINTS } else { "1000" }

docker compose cp scripts/seed_demo_grant.sql postgres:/tmp/acgw-demo-grant.sql | Out-Null
docker compose exec -T postgres psql -U acgw -d acgw -v ON_ERROR_STOP=1 `
    -v "discord_id=$($env:ACGW_DEMO_DISCORD_ID)" -v "points=$points" -f /tmp/acgw-demo-grant.sql
Write-Output "demo: granted $points points to Discord user $($env:ACGW_DEMO_DISCORD_ID)"
