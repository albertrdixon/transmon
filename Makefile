PROJECT = github.com/albertrdixon/transmon
TEST_COMMAND = godep go test
EXECUTABLE = transmon
PKG = .
LDFLAGS = -s
PLATFORMS = linux darwin

.PHONY: dep-save dep-restore test test-verbose build install clean

all: test build

help:
	@echo "Available targets:"
	@echo ""
	@echo "  dep-save"
	@echo "  dep-restore"
	@echo "  test"
	@echo "  test-verbose"
	@echo "  build"
	@echo "  package"
	@echo "  install"
	@echo "  clean"

dep-save:
	@echo "--> Saving dependencies..."
	@godep save ./...

dep-restore:
	@echo "--> Restoring dependencies..."
	@godep restore

test:
	@echo "--> Running all tests"
	@echo ""
	@$(TEST_COMMAND) ./...

test-verbose:
	@echo "--> Running all tests (verbose output)"
	@ echo ""
	@$(TEST_COMMAND) -test.v ./...

build:
	@echo "--> Building executables"
	@GOOS=linux CGO_ENABLED=0 godep go build -a -installsuffix cgo -ldflags '$(LDFLAGS)' -o bin/$(EXECUTABLE)-linux $(PKG)
	@GOOS=darwin CGO_ENABLED=0 godep go build -a -ldflags '$(LDFLAGS)' -o bin/$(EXECUTABLE)-darwin $(PKG)

install:
	@echo "--> Installing..."
	@godep go install ./...

package: build
	@for p in $(PLATFORMS) ; do \
		echo "--> Tar'ing up $$p/amd64 binary" ; \
		test -f bin/$(EXECUTABLE)-$$p && \
		mv bin/$(EXECUTABLE)-$$p $(EXECUTABLE) && \
		tar czf $(EXECUTABLE)-$$p.tgz $(EXECUTABLE) ; \
	done

clean:
	@echo "--> Cleaning up workspace..."
	@go clean ./...
	@rm -rf transmon-* transmon.tgz