.PHONY: build test lint clean

BINARY := oteldoctor
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "0.1.0-dev")
LDFLAGS := -ldflags "-X github.com/firfircelik/oteldoctor/internal/cli.version=$(VERSION)"

build:
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/oteldoctor

test:
	go test -race -count=1 ./...

lint:
	@echo "lint: not implemented"

clean:
	rm -rf bin/
