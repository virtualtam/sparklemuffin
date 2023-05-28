BUILD_DIR ?= build
SRC_FILES := $(shell find . -name "*.go")

all: lint cover build
.PHONY: all

build: $(BUILD_DIR)/sparklemuffin

$(BUILD_DIR)/%: $(SRC_FILES)
	go build -trimpath -o $@ ./cmd/$*

lint:
	golangci-lint run ./...
.PHONY: lint

cover:
	go test -coverprofile=coverage.out ./...
.PHONY: cover

coverhtml: cover
	go tool cover -html=coverage.out
.PHONY: coverhtml

test:
	go test ./...
.PHONY: test

psql:
	@PGPASSWORD=sparklemuffin psql -h localhost -p 15432 -U sparklemuffin
.PHONY: psql

live:
	@echo "== Starting database"
	docker compose -f docker-compose.dev.yml up --remove-orphans -d
	@echo "== Watching for changes... (hit Ctrl+C when done)"
	@watchexec --restart --exts css,go,gohtml -- go run ./cmd/sparklemuffin/ run
.PHONY: live
