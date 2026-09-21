#!/usr/bin/env sh
# Decrypt secrets/<environment>.sops.env into the runtime .env file.
set -eu

environment="${1:-development}"
target="${2:-.env}"
source_file="secrets/${environment}.sops.env"

if [ ! -f "$source_file" ]; then
    echo "missing $source_file" >&2
    exit 1
fi

umask 077
sops decrypt --input-type dotenv --output-type dotenv "$source_file" > "$target"
echo "secrets: wrote $target from $source_file"
