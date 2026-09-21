#!/usr/bin/env sh
# Fail when a plaintext secret or an age private key is tracked by git.
set -eu

cd "$(git rev-parse --show-toplevel)"

status=0

# 1. Every tracked file under secrets/ must be an encrypted set or a template.
for file in $(git ls-files secrets); do
    case "$file" in
        *.env.example | *.sops.env) ;;
        *)
            echo "plaintext secret tracked: $file" >&2
            status=1
            ;;
    esac
done

# 2. An encrypted set must carry SOPS metadata.
for file in $(git ls-files 'secrets/*.sops.env'); do
    if ! grep -q 'ENC\[' "$file" 2>/dev/null; then
        echo "not encrypted (missing SOPS metadata): $file" >&2
        status=1
    fi
done

# 3. No tracked file may contain an age private key header.
if git grep -I -l 'AGE-SECRET-KEY[-]1' -- . >/dev/null 2>&1; then
    git grep -I -l 'AGE-SECRET-KEY[-]1' -- . >&2
    echo "an age private key is tracked (see the files above)" >&2
    status=1
fi

if [ "$status" -eq 0 ]; then
    echo "secrets: no plaintext secret tracked"
fi
exit "$status"
