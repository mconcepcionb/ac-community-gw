# Run the full local demo: prepare the data, start the fake AzerothCore in the
# background, then run the gateway and the SPA (task dev) in the foreground.
$ErrorActionPreference = "Stop"

Set-Location (git rev-parse --show-toplevel)

if (-not (Test-Path .env)) {
    throw "demo: missing .env - configure Discord first (see docs/runbooks/discord-oauth-setup.md)"
}

$envText = Get-Content -Raw .env
$missing = @()
foreach ($key in 'ACGW_DISCORD_CLIENT_ID', 'ACGW_DISCORD_CLIENT_SECRET', 'ACGW_DISCORD_REDIRECT_URL', 'ACGW_DISCORD_GUILD_ID') {
    if ($envText -notmatch "(?m)^$key=..*") { $missing += $key }
}
if ($missing.Count -gt 0) {
    throw "demo: missing required Discord config in .env: $($missing -join ', ') - see docs/runbooks/discord-oauth-setup.md"
}

task demo:setup

New-Item -ItemType Directory -Force .demo | Out-Null
Write-Output "demo: starting the fake AzerothCore on http://localhost:7878/"
$fake = Start-Process -FilePath "go" `
    -ArgumentList "run", "./cmd/fakeazerothcore", "-addr", "127.0.0.1:7878", "-user", "acgw", "-pass", "acgw" `
    -RedirectStandardOutput ".demo/fake-ac.log" -RedirectStandardError ".demo/fake-ac.err.log" `
    -PassThru -NoNewWindow

try {
    Start-Sleep -Seconds 2
    Write-Output ""
    Write-Output "==================================================================="
    Write-Output " demo addresses"
    Write-Output "   SPA (login with Discord): http://localhost:5173/"
    Write-Output "   fake AzerothCore dash:    http://localhost:7878/"
    Write-Output "   walkthrough:              docs/demo.md"
    Write-Output "==================================================================="
    Write-Output ""
    task dev
}
finally {
    if ($fake -and -not $fake.HasExited) {
        Stop-Process -Id $fake.Id -Force
    }
}
