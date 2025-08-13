# Copyright (c) VirtualTam
# SPDX-License-Identifier: MIT

# Step 1: Build frontend assets
FROM node:24-bookworm AS assets

WORKDIR /app
COPY internal/http/www/assets/package.json internal/http/www/assets/package-lock.json ./
RUN --mount=type=cache,target=/root/.npm npm ci

# Step 2: Build Go binaries
FROM golang:1.24-bookworm AS builder

ARG CGO_ENABLED=1

WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

ADD . .
COPY --from=assets /app/node_modules internal/http/www/assets/node_modules
RUN --mount=type=cache,target=/root/.cache/go-build make build

# Step 3: Build the final image
FROM debian:bookworm-slim

RUN --mount=type=cache,target=/var/lib/apt/lists \
    --mount=type=cache,target=/var/cache/apt \
    rm -f /etc/apt/apt.conf.d/docker-clean \
    && apt update \
    && apt install -y ca-certificates

RUN groupadd \
        --gid 1000 \
        sparklemuffin \
    && useradd \
        --create-home \
        --home-dir /var/lib/sparklemuffin \
        --shell /bin/bash \
        --uid 1000 \
        --gid sparklemuffin \
        sparklemuffin

COPY --from=builder /app/build/sparklemuffin /usr/local/bin/sparklemuffin

ENV \
    SPARKLEMUFFIN_CSRF_KEY="csrf-secret-key" \
    SPARKLEMUFFIN_DB_ADDR="postgres:5432" \
    SPARKLEMUFFIN_DB_SSLMODE="disable" \
    SPARKLEMUFFIN_DB_NAME="sparklemuffin" \
    SPARKLEMUFFIN_DB_USER="sparklemuffin" \
    SPARKLEMUFFIN_DB_PASSWORD="sparklemuffin" \
    SPARKLEMUFFIN_HMAC_KEY="hmac-secret-key" \
    SPARKLEMUFFIN_LISTEN_ADDR="0.0.0.0:8080" \
    SPARKLEMUFFIN_METRICS_LISTEN_ADDR="127.0.0.1:8081" \
    SPARKLEMUFFIN_PUBLIC_ADDR="http://localhost:8080" \
    SPARKLEMUFFIN_LOG_FORMAT="json" \
    SPARKLEMUFFIN_LOG_LEVEL="info"

EXPOSE 8080

USER sparklemuffin
WORKDIR /var/lib/sparklemuffin

CMD ["/usr/local/bin/sparklemuffin", "run"]
