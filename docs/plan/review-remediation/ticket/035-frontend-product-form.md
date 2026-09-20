# 035 — Product form correctness

**Phase:** 6 · **Gate:** G-standard, contributes to G6-frontend · **Depends on:** 001

## Goal

Make the product edit dialog operate on the correct resource and reset when the
selected product changes.

## Findings addressed

- Medium: `product-form-dialog.tsx` uses the edited `values.sku` as the update
  path instead of the original `product.sku`; renaming a SKU issues the request
  against a non-existent resource and leaves stale cache entries.
- Medium: `useForm({ defaultValues: initialValues(product) })` reads `product`
  only on first mount; TanStack Router reuses the component across param changes,
  so the dialog shows the previous product's values. The same pattern appears in
  `set-email-dialog.tsx` and `set-password-dialog.tsx`.

## Context

`web/src/features/store/product-form-dialog.tsx`,
`accounts/set-email-dialog.tsx`, `accounts/set-password-dialog.tsx`.

## Atomic change

Capture the original key for the request/invalidation and reset/remount forms
when props change.

## Requirements

- Use the original resource key for the path and cache invalidation
  (`mode === "update" ? product.sku : values.sku`); update the edit form to show
  the original SKU as read-only or clearly separate "rename" from "edit".
- Reset forms when their entity changes (`useEffect(() => form.reset(...), [product, form])`)
  or key dialogs by identity.
- Apply the same reset-on-prop-change pattern to the account dialogs.
- Invalidate the correct list/detail query keys after a successful update.
- Keep existing validation and error display.

## Tests

- Test: editing without changing the SKU calls the original SKU path.
- Test: renaming the SKU calls the original SKU path and invalidates both old and
  new keys.
- Test: selecting product A then B shows B's values (no stale form).
- Apply the same assertions to the account dialogs.

## Acceptance criteria (gate G-standard, G6-frontend)

- `task web:check` green.
- Update requests always target the resource that was opened.

## Rollback

Revert; the stale-form and wrong-path bugs return.

## Out of scope

- The catalog list rendering (ticket 037).
