# OpenAPI client codegen

## Goal

Generate a typed TypeScript client from `api/swagger.yaml` and wire the runtime
HTTP client, including error normalisation.

## Context

The spec is produced incrementally from tickets 001..018. The client grows with
it; only annotated operations exist in generated code.

## Requirements

- `web/openapi-ts.config.ts` using `@hey-api/openapi-ts` with plugins:
  `@hey-api/client-fetch`, `@hey-api/typescript`, `@hey-api/sdk`,
  `@tanstack/react-query`, `zod`; input `../api/swagger.yaml`; output
  `src/api/generated/`.
- Generated code committed under `web/src/api/generated/`.
- `src/api/client.ts`: configure the generated fetch client with `baseUrl`
  (from `import.meta.env`), `credentials: 'include'`, and an interceptor that
  maps the backend error envelope
  `{ error: { code, message, details? }, request_id? }` to a typed `ApiError`.
- `src/api/errors.ts`: `ApiError` with `status`, `code`, `message`, `details`,
  `requestId`, plus an `isUnauthorized(err)` helper.
- TypeScript types re-exported from `src/api/types.ts` for feature code.
- Taskfile: `openapi:client`, `openapi` (spec + client), and extend
  `openapi:check` to also verify the generated client (`git diff --exit-code
  web/src/api/generated`).

## Acceptance criteria

- `task openapi` regenerates spec and client deterministically.
- Generated client includes the currently annotated operations (auth, health).
- `task web:check` green; `task openapi:check` red after a manual client edit.
- ApiError mapping unit-tested with representative envelopes (401, 403, 422
  with details, 500 without request id).

## Implementation notes

- Pin `@hey-api/openapi-ts` to an exact version (pre-1.0).
- Configure `operationId`-derived names; keep `@ID`s stable upstream.
- Do not hand-edit anything under `src/api/generated/`.

## Tests

- Vitest + MSW: a generated SDK call returns data; a 4xx envelope throws an
  `ApiError` with the right `code` and `status`.

## Dependencies

- 001, 019. Unblocks 021..034.
