REV ?= $$(git rev-parse --short=8 HEAD)
BRANCH ?= $$(git rev-parse --abbrev-ref HEAD | tr / _)
EXECUTABLE = transmon
LDFLAGS = "-linkmode external -extldflags '-static -s'"
TEST_COMMAND = godep go test

.PHONY: dep-save dep-restore test test-verbose build build-image install publish

all: test build install
container: build-image publish

help:
	@echo "Available targets:"
	@echo ""
	@echo "  dep-save       : Save dependencies (godep save)"
	@echo "  dep-restore    : Restore dependencies (godep restore)"
	@echo "  test           : Run package tests"
	@echo "  test-verbose   : Run package tests with verbose output"
	@echo "  build          : Build binary (go build)"
	@echo "  build-image    : Build binary and container image"
	@echo "  install        : Install binary (go install)"
	@echo "  publish        : Publish container image to remote repo"

dep-save:
	@echo "==> Saving dependencies to ./Godeps"
	@godep save -t -v ./...

dep-restore:
	@echo "==> Restoring dependencies from ./Godeps"
	@godep restore -v

test:
	@echo "==> Running all tests"
	@echo ""
	@$(TEST_COMMAND) ./...

test-verbose:
	@echo "==> Running all tests (verbose output)"
	@echo ""
	@$(TEST_COMMAND) -test.v ./...

build:
	@echo "==> Building $(EXECUTABLE) with ldflags '$(LDFLAGS)'"
	@godep go build -ldflags $(LDFLAGS) -o bin/$(EXECUTABLE) *.go

install:
	@echo "==> Installing $(EXECUTABLE) with ldflags '$(LDFLAGS)'"
	@godep go install -ldflags $(LDFLAGS) $(BINARY)