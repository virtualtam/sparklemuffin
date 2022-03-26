BUILD_DIR ?= build
SRC_FILES := $(shell find . -name "*.go")

all: lint cover build
.PHONY: all

build: $(BUILD_DIR)/yawbe-srv

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

live:
	@echo "== Watching for changes... (hit Ctrl+C when done)"
	@watchexec --restart --exts css,go,gohtml -- go run ./cmd/yawbe-srv/
.PHONY: live
