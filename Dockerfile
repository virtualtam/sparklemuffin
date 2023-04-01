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
        yawbe \
    && useradd \
        --create-home \
        --home-dir /var/lib/yawbe \
        --shell /bin/bash \
        --uid 1000 \
        --gid yawbe \
        yawbe

COPY --from=builder /app/build/yawbe /usr/local/bin/yawbe

ENV \
    YAWBE_DB_ADDR="postgres:5432" \
    YAWBE_DB_NAME="yawbe" \
    YAWBE_DB_USER="yawbe" \
    YAWBE_DB_PASSWORD="yawbe" \
    YAWBE_HMAC_KEY="hmac-secret-key" \
    YAWBE_LISTEN_ADDR="0.0.0.0:8080" \
    YAWBE_PUBLIC_ADDR="http://localhost:8080" \
    YAWBE_LOG_LEVEL="info"

EXPOSE 8080

USER yawbe
WORKDIR /var/lib/yawbe

CMD ["/usr/local/bin/yawbe", "run"]
