# syntax=docker/dockerfile:1

# SPA image: builds the frontend and serves it with Caddy, which also
# reverse-proxies /api/* and the operational endpoints to the gateway.
# Build from the repository root:
#   docker build -f web/Dockerfile -t ac-community-gw-web:local .

FROM node:lts-alpine AS build
WORKDIR /src
RUN npm install -g pnpm@12.4.1
COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile
COPY web/ ./
RUN pnpm build

FROM caddy:2-alpine

COPY web/Caddyfile /etc/caddy/Caddyfile
COPY --from=build /src/dist /srv

USER caddy

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=5 \
    CMD wget -q -O /dev/null http://127.0.0.1:8080/ || exit 1
