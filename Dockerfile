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

RUN mkdir /opt/yawbe
WORKDIR /opt/yawbe

COPY --from=builder /app/build/yawbe-srv /opt/yawbe/yawbe-srv

ENV \
    YAWBE_DB_ADDR="postgres:5432" \
    YAWBE_DB_NAME="yawbe" \
    YAWBE_DB_USER="yawbe" \
    YAWBE_DB_PASSWORD="yawbe"

CMD ["/opt/yawbe/yawbe-srv", "run"]
