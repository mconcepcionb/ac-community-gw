# Runbook: rotate the Discord client secret

## Purpose

Replace the Discord OAuth2 client secret without downtime or data loss.

## Preconditions

- Access to the Discord developer portal.
- Ability to update runtime environment variables and restart the service.

## Procedure

1. In the Discord developer portal, generate a new client secret.
2. Update `ACGW_DISCORD_CLIENT_SECRET` in the runtime environment or secret
   store. Never commit it.
3. Restart the gateway.
4. Revoke the old secret in the portal.

## Validation

- `GET /api/v1/auth/discord/login` still redirects to Discord.
- A login completes the callback flow.

## Rollback

If logins fail, restore the previous secret value temporarily, restart, and
investigate before revoking it.

## Troubleshooting

- Token exchange failures usually indicate a mismatched client id/secret or
  redirect URI.
- Never copy secrets into tickets, logs or chat.
