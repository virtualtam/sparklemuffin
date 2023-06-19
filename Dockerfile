# Step 1: Build Go binaries
FROM golang:1.20-bullseye as builder

ARG CGO_ENABLED=1

WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

ADD . .
RUN --mount=type=cache,target=/root/.cache/go-build make build

# Step 2: Build the actual image
FROM debian:bullseye-slim

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
    SPARKLEMUFFIN_DB_PASSWORD="sparklemuffin" \
    SPARKLEMUFFIN_HMAC_KEY="hmac-secret-key" \
    SPARKLEMUFFIN_LISTEN_ADDR="0.0.0.0:8080" \
    SPARKLEMUFFIN_PUBLIC_ADDR="http://localhost:8080" \
    SPARKLEMUFFIN_LOG_LEVEL="info"

EXPOSE 8080

USER sparklemuffin
WORKDIR /var/lib/sparklemuffin

CMD ["/usr/local/bin/sparklemuffin", "run"]
