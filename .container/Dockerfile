# syntax=docker/dockerfile:1

# --- UI build stage -----------------------------------------------------------
FROM node:20-slim AS ui

WORKDIR /app

RUN --mount=type=cache,target=/root/.npm \
    npm install -g pnpm@11.6.0

COPY ui/package.json ui/pnpm-lock.yaml ui/pnpm-workspace.yaml ./ui/
RUN --mount=type=cache,target=/root/.local/share/pnpm/store \
    pnpm --prefix ui install --frozen-lockfile

COPY ui ./ui
RUN pnpm --prefix ui build

# --- Go build stage -----------------------------------------------------------
FROM golang:1.26 AS build

ENV CGO_ENABLED=1
ENV GOOS=linux

ARG BUILD_DATE=
ARG VERSION=docker

WORKDIR /app

RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    apt-get update && apt-get install -y --no-install-recommends libzmq3-dev

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
COPY --from=ui /app/ui/dist ./internal/web/dist

RUN mkdir build

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    BUILD_DATE="${BUILD_DATE:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}" && \
    go build -ldflags="-s -w -X main.Version=${VERSION} -X main.DateBuilt=${BUILD_DATE}" -o build ./cmd/...

# --- Runtime stage ------------------------------------------------------------
FROM debian:bookworm-slim AS runtime

RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    apt-get update && apt-get install -y --no-install-recommends \
    libzmq3-dev \
    ca-certificates \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

COPY --from=build /app/build .

CMD ["/qaynaq", "run"]
