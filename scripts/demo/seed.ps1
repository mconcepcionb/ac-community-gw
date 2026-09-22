# Seed the demo data. Idempotent; safe to re-run.
$ErrorActionPreference = "Stop"

Set-Location (git rev-parse --show-toplevel)

function Invoke-PsqlFile {
    param([string]$File, [string[]]$PsqlArgs = @())
    docker compose cp $File postgres:/tmp/acgw-demo.sql | Out-Null
    docker compose exec -T postgres psql -U acgw -d acgw -v ON_ERROR_STOP=1 @PsqlArgs -f /tmp/acgw-demo.sql
}

Write-Output "demo: seeding the store catalog"
Invoke-PsqlFile -File scripts/seed_demo_store.sql

if ($env:ACGW_DEMO_DISCORD_ROLE_ID) {
    Write-Output "demo: mapping Discord role $($env:ACGW_DEMO_DISCORD_ROLE_ID) to the internal role ac-core.admin"
    Invoke-PsqlFile -File scripts/seed_demo_admin.sql `
        -PsqlArgs @("-v", "discord_role_id=$($env:ACGW_DEMO_DISCORD_ROLE_ID)", "-v", "internal_role=ac-core.admin")
}
else {
    Write-Output "demo: ACGW_DEMO_DISCORD_ROLE_ID is not set; skipped the admin role mapping"
    Write-Output "demo: set it in .env to give the demo account staff permissions"
}
