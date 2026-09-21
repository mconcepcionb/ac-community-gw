# Decrypt one encrypted secret set to stdout.
param([Parameter(Mandatory)][string]$Path)

sops decrypt --input-type dotenv --output-type dotenv $Path
