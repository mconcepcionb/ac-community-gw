# Runbook: debug AzerothCore connectivity

## Purpose

Diagnose failures between the gateway and the AzerothCore SOAP endpoint.

## Preconditions

- Access to gateway logs and runtime configuration.
- Network access to the SOAP endpoint from the gateway host/container.

## Procedure

1. Confirm configuration:

```text
ACGW_AZEROTH_SOAP_URL
ACGW_AZEROTH_SOAP_USERNAME
ACGW_AZEROTH_SOAP_PASSWORD
```

2. Confirm the process is not logging `azerothcore: command executor not
   configured` (which means no URL is set).
3. From the gateway network, test reachability with a controlled client (for
   example the gateway's own info route).
4. Inspect logs for transport errors: timeout, connection refused, 401, or SOAP
   fault.

## Validation

- A read-only capability succeeds and returns command output.
- `/readyz` is 200.

## Rollback

No state is changed by diagnostics.

## Troubleshooting

- `ErrNotConfigured`: the URL is empty.
- `ErrHTTPStatus` 401: credentials rejected.
- timeout: network path or SOAP timeout too low.
- `ErrSOAPFault`: AzerothCore rejected the command itself.
- Never add a public generic command endpoint for debugging.
