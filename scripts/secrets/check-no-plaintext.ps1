# Fail when a plaintext secret or an age private key is tracked by git.
$ErrorActionPreference = "Stop"

Set-Location (git rev-parse --show-toplevel)

$status = 0

# 1. Every tracked file under secrets/ must be an encrypted set or a template.
foreach ($file in git ls-files secrets) {
    if ($file -match '\.env\.example$' -or $file -match '\.sops\.env$') { continue }
    Write-Error "plaintext secret tracked: $file"
    $status = 1
}

# 2. An encrypted set must carry SOPS metadata.
foreach ($file in git ls-files 'secrets/*.sops.env') {
    $content = Get-Content -Raw -LiteralPath $file
    if ($content -notmatch 'ENC\[') {
        Write-Error "not encrypted (missing SOPS metadata): $file"
        $status = 1
    }
}

# 3. No tracked file may contain an age private key header.
$ageHits = git grep -I -l 'AGE-SECRET-KEY[-]1' -- .
if ($LASTEXITCODE -eq 0 -and $ageHits) {
    $ageHits | ForEach-Object { Write-Error "age private key tracked: $_" }
    $status = 1
}

if ($status -eq 0) { Write-Output "secrets: no plaintext secret tracked" }
exit $status
