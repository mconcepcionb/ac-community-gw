# 0013 Decoupled portal and console surfaces

## Status

Accepted.

Supersedes the single-surface layout described in
[docs/plan/spa/README.md](../plan/spa/README.md).

## Context

The SPA began as one flat console: a single shell whose navigation listed every
resource page, gated by permission. It served two very different audiences with
one layout - community players and staff - and offered no place for
self-service onboarding, so account linking was an administrator-only action.

The desired use cases ([docs/use-cases.md](../use-cases.md)) call for a
player-facing portal and a staff console, each with its own navigation, on one
origin.

## Decision

Split the SPA into two surfaces:

- a **player portal** at `/`, and
- a **staff console** at `/admin/*`.

Both share one origin, one session cookie and one backend. The root route is a
minimal document wrapper; each surface is a layout route that owns its own
navigation, so the two never share a chrome.

Landing is permission-driven: after sign-in, a user holding any console
permission lands on `/admin`, everyone else on `/`. A user who can only use the
portal is forbidden (not redirected in a loop) from `/admin`.

Existing flat paths were **removed, not redirected**. This is deliberate: the
API is still evolving and carrying aliases would freeze the old information
architecture. The change is documented per route in
[docs/frontend.md](../frontend.md).

The serving model is unchanged: the gateway never serves the SPA
([ADR 0012](0012-decoupled-spa-serving.md)).

## Consequences

- Each surface builds its navigation from `/me` permissions and hides what the
  user cannot open.
- New self-service endpoints back the portal: account onboarding and claims,
  self-scoped characters and mail, per-character public visibility, and the
  storefront and wallet.
- Cross-cutting reads the console needs (the user 360 view, all orders) are
  exposed as dedicated staff endpoints rather than reusing self-scoped ones.
- Bookmarks to retired paths break; the landing resolver makes the new
  structure discoverable.
