# 036 — Dialog and mutation correctness

**Phase:** 6 · **Gate:** G-standard, contributes to G6-frontend · **Depends on:** 001

## Goal

Stop confirm dialogs from closing on validation failure, and invalidate the right
queries after mutations.

## Findings addressed

- Medium: `confirm-dialog.tsx` closes when the confirm callback resolves, but
  `react-hook-form`'s `handleSubmit` resolves even when validation fails, so the
  dialog closes and hides the error (affects `announce-form`, `purchase-dialog`,
  `grant-dialog`, `mail-form`).
- Medium: `mail-form.tsx` calls `invalidateQueries()` with no key, refetching the
  entire cache after a send.
- Low: `grant-dialog.tsx` invalidates `storeWalletGetQueryKey()` (the current
  user's wallet), not the target user's wallet.

## Context

`web/src/components/common/confirm-dialog.tsx`, `mutation-button.tsx`, and the
four forms. TanStack Query keys are generated.

## Atomic change

Make confirmation await an explicit success signal and scope invalidations to the
affected resources.

## Requirements

- `ConfirmDialog`: keep the dialog open when the async action reports failure;
  close only on success. Have callers return a boolean/throw so the component can
  decide, or pass the RHF valid-submit result through.
- `mail-form`: invalidate only the mail/character query keys that change.
- `grant-dialog`: invalidate the target user's wallet key (by `user_id`/
  `discord_id`), or drop the no-op invalidation.
- Catch/display rejections so no unhandled promise rejection escapes.
- Audit the other dialogs (`ban`, `set-gmlevel`, `purchase`, `announce`) for the
  same pattern.

## Tests

- Test: confirm with invalid form keeps the dialog open and shows the error.
- Test: success closes the dialog.
- Test: a failed mutation rejection is surfaced, not unhandled.
- Test: mail send invalidates only the expected keys.
- Test: grant invalidates the target wallet key.

## Acceptance criteria (gate G-standard, G6-frontend)

- `task web:check` green.
- No dialog closes while its form has validation errors.

## Rollback

Revert; the premature-close and over-invalidation return.

## Out of scope

- Reworking the mutation-button UX.
