BUILD_DIR ?= build
SRC_FILES := $(shell find . -name "*.go")

all: lint race cover build
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

race:
	go test -race ./...
.PHONY: race

test:
	go test ./...
.PHONY: test

vulncheck:
	govulncheck -C . ./...
.PHONY: vulncheck

# Live development server
live:
	@echo "== Starting database"
	docker compose -f docker-compose.dev.yml up --remove-orphans -d
	@echo "== Watching for changes... (hit Ctrl+C when done)"
	@watchexec --restart --exts css,go,gohtml -- go run ./cmd/sparklemuffin/ run
.PHONY: live

# Live development server (with race detection enabled)
live-race:
	@echo "== Starting database"
	docker compose -f docker-compose.dev.yml up --remove-orphans -d
	@echo "== Watching for changes... (hit Ctrl+C when done)"
	@watchexec --restart --exts css,go,gohtml -- go run -race ./cmd/sparklemuffin/ run
.PHONY: live-race

# Live development server - PostgreSQL console
psql:
	@PGPASSWORD=sparklemuffin psql -h localhost -p 15432 -U sparklemuffin
.PHONY: psql

# Live development server - Database migrations
dev-migrate:
	go run ./cmd/sparklemuffin migrate
.PHONY: dev-migrate

# Live development server - Synchronize feeds
dev-sync-feeds:
	go run ./cmd/sparklemuffin sync-feeds
.PHONY: dev-feed-sync

# Live development server - Create administrator user
dev-admin:
	go run ./cmd/sparklemuffin createadmin \
		--displayname Admin \
		--email admin@dev.local \
		--nickname admin
.PHONY: dev-admin

# Documentation
DOCS_DIR := docs
DOCS_FILES := $(shell find docs -name "*.md")

docs: docs/book
.PHONY: docs

docs/book: $(DOCS_FILES)
	mdbook build $(DOCS_DIR)

live-docs:
	mdbook serve $(DOCS_DIR)
.PHONY: live-docs
