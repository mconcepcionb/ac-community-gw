# Runbook: rotate the Discord client secret

## Purpose

Replace the Discord OAuth2 client secret without downtime or data loss.

## Preconditions

- Access to the Discord developer portal.
- Access to the encrypted secret set (`secrets/<environment>.sops.env`) and an
  age identity allowed to decrypt it (`SOPS_AGE_KEY_FILE`, see
  [secret-management.md](secret-management.md)).

## Procedure

1. In the Discord developer portal, generate a new client secret.
2. Edit the encrypted set and update `ACGW_DISCORD_CLIENT_SECRET`:

   ```sh
   export SOPS_AGE_KEY_FILE=~/.config/ac-community-gw/age/admin.key.txt
   task secrets:edit ENV=development      # or ENV=production on the deploy host
   ```

3. Materialize the set where the gateway runs and restart it:

   ```sh
   task secrets:decrypt ENV=development   # writes .env
   ```

4. Revoke the old secret in the portal.

The previously exposed development secret must be rotated this way and the old
value revoked; see the [secrets plan](../plan/secrets/README.md) ticket S6.

## Validation

- `GET /api/v1/auth/discord/login` still redirects to Discord.
- A login completes the callback flow.

## Rollback

If logins fail, restore the previous value in the encrypted set
(`task secrets:edit`), materialize, restart, and investigate before revoking it.

## Troubleshooting

- Token exchange failures usually indicate a mismatched client id/secret or
  redirect URI.
- Never copy secrets into tickets, logs or chat, and never paste a secret into
  a plaintext file outside the encrypted set.
