# 032 — Container hardening

**Phase:** 5 · **Gate:** G-standard, contributes to G5-infra · **Depends on:** 029

## Goal

Harden the images: run the SPA proxy as non-root, add healthchecks, pin base
images, and wire compose dependencies to health.

## Findings addressed

- Medium/High: `web/Dockerfile` runs Caddy as root and has no `USER`; neither
  image has a healthcheck; base images (`golang:1.27`, `distroless:nonroot`,
  `node:lts-alpine`, `caddy:2-alpine`, `cloudflare/cloudflared:latest`) are not
  pinned by digest.
- Low: `compose.yaml` depends on `service_started` only, so Caddy can serve `502`
  until the gateway is healthy; no app healthcheck.

## Context

`Dockerfile`, `web/Dockerfile`, `compose.yaml`. The Go image is already
multi-stage, static, `CGO_ENABLED=0`, and runs as `nonroot:nonroot`.

## Atomic change

Pin digests, run Caddy as non-root, add healthchecks, and switch compose
dependencies to `service_healthy`.

## Requirements

- Pin every base image by `@sha256:` digest; document a Dependabot/Renovate policy
  to bump them.
- `web/Dockerfile`: add `USER caddy` (Caddy listens on unprivileged `8080`); ensure
  file permissions allow non-root binding.
- Add `HEALTHCHECK` for the gateway (distroless has no shell — use the binary or
  an orchestrator probe) and for the SPA.
- `compose.yaml`: `depends_on` with `condition: service_healthy` for `web → app`
  and `app → db`.
- Keep the Go image multi-stage/static and non-root.

## Tests

- `docker build` both images and inspect `User`/`Healthcheck`.
- `docker compose up` shows healthy dependencies and no 502 window.
- Image scan (Trivy/Grype) reports no Critical/High after pinning.

## Acceptance criteria (gate G-standard, G5-infra)

- Both images run as non-root; both have healthchecks.
- All base images are digest-pinned.
- Compose starts in dependency order with health gating.

## Rollback

Revert the Dockerfiles/compose; the root/healthcheck issues return.

## Out of scope

- Distroless hardening of the SPA image beyond non-root.
- CI registry publishing.
