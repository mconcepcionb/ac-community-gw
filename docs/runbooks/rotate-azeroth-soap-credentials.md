# Runbook: rotate AzerothCore SOAP credentials

## Purpose

Replace the SOAP username/password used to talk to AzerothCore.

## Preconditions

- Access to the AzerothCore SOAP configuration.
- Ability to update runtime environment variables and restart the gateway.

## Procedure

1. Create the new SOAP account/credentials in AzerothCore.
2. Verify the new credentials work (see validation).
3. Update `ACGW_AZEROTH_SOAP_USERNAME` and `ACGW_AZEROTH_SOAP_PASSWORD`.
4. Restart the gateway.
5. Remove the old credentials from AzerothCore.

## Validation

- `GET /readyz` is 200.
- Trigger a safe read-only capability, for example the info status route, and
  confirm a successful SOAP response.

## Rollback

Restore the previous credentials, restart, and investigate before removing the
old account.

## Troubleshooting

- 401 from SOAP: wrong credentials or account lacks command permission.
- Connection errors: check the SOAP URL and network reachability.
- Never expose the SOAP endpoint publicly.
