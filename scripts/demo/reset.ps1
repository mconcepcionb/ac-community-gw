# Reset the demo between runs: clear the fake AzerothCore state and re-seed.
$ErrorActionPreference = "Stop"

Set-Location (git rev-parse --show-toplevel)

try {
    Invoke-RestMethod -Method Post -Uri http://localhost:7878/reset -TimeoutSec 5 | Out-Null
    Write-Output "demo: reset the fake AzerothCore state"
}
catch {
    Write-Output "demo: fake AzerothCore is not running (start it with task demo)"
}

& pwsh -NoProfile -File scripts/demo/seed.ps1
