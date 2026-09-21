# Decrypt secrets/<environment>.sops.env into the runtime .env file.
param(
    [string]$Environment = "development",
    [string]$Target = ".env"
)

$source_file = "secrets/$Environment.sops.env"
if (-not (Test-Path -LiteralPath $source_file)) {
    throw "missing $source_file"
}

$decrypted = sops decrypt --input-type dotenv --output-type dotenv $source_file
$content = ($decrypted -join "`n") + "`n"
[System.IO.File]::WriteAllText((Join-Path (Get-Location) $Target), $content)
Write-Output "secrets: wrote $Target from $source_file"
