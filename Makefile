BUILD_DIR ?= build
SRC_FILES := $(shell find . -name "*.go")
POSTGRESQL_FILES = internal/repository/postgresql/migrations

all: lint race cover build docs
.PHONY: all

build: $(BUILD_DIR)/sparklemuffin

$(BUILD_DIR)/%: $(SRC_FILES)
	go build -trimpath -o $@ ./cmd/$*

lint:
	golangci-lint run ./...
.PHONY: lint

format: copywrite format-sql
.PHONY: format

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

format-sql:
	sqlfluff format $(POSTGRESQL_FILES)
.PHONY: format-sql

lint-sql:
	sqlfluff lint --disable-progress-bar $(POSTGRESQL_FILES)
.PHONY: lint-sql

# Install development tools
dev-install-tools:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.64.5
	go install github.com/hashicorp/copywrite@latest
	go install golang.org/x/tools/gopls@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
.PHONY: dev-install-tools

dev-install-sqlfluff:
	pip install 'sqlfluff==3.2.5'
.PHONY: dev-install-sqlfluff

# Licence headers
copywrite:
	copywrite headers
.PHONY: copywrite

# Modernize
modernize:
	go run golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@latest -fix -test ./...
.PHONY: modernize

# Vulnerability check
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
	docker compose exec postgres psql -U sparklemuffin
.PHONY: psql

# Live development server - Database migrations
dev-migrate:
	go run ./cmd/sparklemuffin migrate
.PHONY: dev-migrate

# Live development server - Synchronize feeds
dev-sync-feeds:
	go run ./cmd/sparklemuffin sync-feeds
.PHONY: dev-sync-feeds

# Live development server - Create administrator user
dev-admin:
	go run ./cmd/sparklemuffin createadmin \
		--displayname Admin \
		--email admin@dev.local \
		--nickname admin
.PHONY: dev-admin

# Documentation
DOCS_DIR := docs
DOCS_FILES := $(shell find docs -name "*.md" -or -name "*.toml")

docs: docs/book/html
.PHONY: docs

docs/book/html: $(DOCS_FILES)
	mdbook build $(DOCS_DIR)

live-docs:
	mdbook serve $(DOCS_DIR)
.PHONY: live-docs
