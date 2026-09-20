# syntax=docker/dockerfile:1

# The API image carries only the gateway. The SPA is built and served
# separately by web/Dockerfile (Caddy). See docs/ADR/0012.

FROM golang:1.27 AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/ac-community-gw \
    ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/ac-community-gw /ac-community-gw

USER nonroot:nonroot

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=5 \
    CMD ["/ac-community-gw", "healthcheck"]

ENTRYPOINT ["/ac-community-gw"]
