#!/usr/bin/env sh
# Decrypt one encrypted secret set to stdout.
set -eu

file="${1:?usage: decrypt-env.sh secrets/<environment>.sops.env}"
exec sops decrypt --input-type dotenv --output-type dotenv "$file"
