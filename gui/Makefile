# Makefile for ham-apps-gui
#
# Requirements:
#   - Go 1.21+
#   - CGO_ENABLED=1 (required for Fyne's OpenGL renderer)
#   - gcc, libGL and libX11 development headers
#
# Targets:
#   build  — compile to ham-apps-gui binary (stripped)
#   test   — go vet + go test
#   lint   — golangci-lint (must be installed separately)
#   check  — test + bash integration tests from ../tests/
#   clean  — remove ham-apps-gui binary

BINARY   := ham-apps-gui
CMD_PKG  := ./cmd/ham-apps-gui

.PHONY: build test lint check clean

build:
	CGO_ENABLED=1 go build -ldflags="-s -w" -o $(BINARY) $(CMD_PKG)

test:
	go vet ./...
	go test ./...

lint:
	golangci-lint run

check: test
	@for t in ../tests/test-*; do \
		echo "Running $$t ..."; \
		bash "$$t" || exit 1; \
	done

clean:
	rm -f $(BINARY)
