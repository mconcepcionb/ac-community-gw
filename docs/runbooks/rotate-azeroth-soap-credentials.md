# Runbook: rotate AzerothCore SOAP credentials

## Purpose

Replace the SOAP username/password used to talk to AzerothCore.

## Preconditions

- Access to the AzerothCore SOAP configuration.
- Access to the encrypted secret set (`secrets/<environment>.sops.env`) and an
  age identity allowed to decrypt it (`SOPS_AGE_KEY_FILE`, see
  [secret-management.md](secret-management.md)).

## Procedure

1. Create the new SOAP account/credentials in AzerothCore.
2. Edit the encrypted set and update `ACGW_AZEROTH_SOAP_USERNAME` and
   `ACGW_AZEROTH_SOAP_PASSWORD`:

   ```sh
   export SOPS_AGE_KEY_FILE=~/.config/ac-community-gw/age/admin.key.txt
   task secrets:edit ENV=development      # or ENV=production on the deploy host
   ```

3. Materialize the set where the gateway runs and restart it:

   ```sh
   task secrets:decrypt ENV=development   # writes .env
   ```

4. Remove the old credentials from AzerothCore.

The previously exposed SOAP password must be rotated this way and the old
credentials removed; see the [secrets plan](../plan/secrets/README.md) ticket S6.

## Validation

- `GET /readyz` is 200.
- Trigger a safe read-only capability, for example the info status route, and
  confirm a successful SOAP response.

## Rollback

Restore the previous credentials in the encrypted set (`task secrets:edit`),
materialize, restart, and investigate before removing the old account.

## Troubleshooting

- 401 from SOAP: wrong credentials or account lacks command permission.
- Connection errors: check the SOAP URL and network reachability.
- Never expose the SOAP endpoint publicly.
