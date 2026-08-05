# Copyright VirtualTam 2022, 2026
# SPDX-License-Identifier: MIT

# Step 1: Build frontend assets
FROM node:24-trixie AS assets

WORKDIR /app
COPY internal/http/www/assets/package.json internal/http/www/assets/package-lock.json ./
RUN --mount=type=cache,target=/root/.npm npm ci

# Step 2: Build Go binaries
FROM golang:1.26-trixie AS builder

ARG CGO_ENABLED=1

WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

ADD . .
COPY --from=assets /app/node_modules internal/http/www/assets/node_modules
RUN --mount=type=cache,target=/root/.cache/go-build make build

# Step 3: Build the final image
FROM debian:trixie-slim

RUN --mount=type=cache,target=/var/lib/apt/lists \
    --mount=type=cache,target=/var/cache/apt \
    rm -f /etc/apt/apt.conf.d/docker-clean \
    && apt update \
    && apt install -y ca-certificates curl

# Add certificates for Let's Encrypt Generation Y CAs (YE and YR).
#
# As of 2026-08-05:
#
# > These roots are not yet included in Root Program Trust Stores, but will be submitted for inclusion soon.
#
# See:
# - https://letsencrypt.org/certificates/
# - https://letsencrypt.org/docs/certificate-compatibility/
# - https://community.letsencrypt.org/t/chain-validation-issues-with-ye-yr-under-linux-distributions/247836
ADD https://letsencrypt.org/certs/gen-y/root-ye.pem /usr/local/share/ca-certificates/le-ye.crt
ADD https://letsencrypt.org/certs/gen-y/int-ye1.pem /usr/local/share/ca-certificates/le-ye1.crt
ADD https://letsencrypt.org/certs/gen-y/int-ye2.pem /usr/local/share/ca-certificates/le-ye2.crt
ADD https://letsencrypt.org/certs/gen-y/root-yr.pem /usr/local/share/ca-certificates/le-yr.crt
ADD https://letsencrypt.org/certs/gen-y/int-yr1.pem /usr/local/share/ca-certificates/le-yr1.crt
ADD https://letsencrypt.org/certs/gen-y/int-yr2.pem /usr/local/share/ca-certificates/le-yr2.crt
RUN update-ca-certificates

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
    SPARKLEMUFFIN_DB_ADDR="postgres:5432" \
    SPARKLEMUFFIN_DB_SSLMODE="disable" \
    SPARKLEMUFFIN_DB_NAME="sparklemuffin" \
    SPARKLEMUFFIN_DB_USER="sparklemuffin" \
    SPARKLEMUFFIN_LOG_FORMAT="json" \
    SPARKLEMUFFIN_LOG_LEVEL="info" \
    SPARKLEMUFFIN_LISTEN_ADDR="0.0.0.0:8080" \
    SPARKLEMUFFIN_MONITORING_LISTEN_ADDR="0.0.0.0:8090" \
    SPARKLEMUFFIN_PUBLIC_ADDR="http://localhost:8080" \
    SPARKLEMUFFIN_CLIENT_IP_HEADER=""

EXPOSE 8080 8090

USER sparklemuffin
WORKDIR /var/lib/sparklemuffin

CMD ["/usr/local/bin/sparklemuffin", "run"]

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8090/health || exit 1
