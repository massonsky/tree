# Makefile for building and running the three CLI

BINARY := three
PKG := ./cmd/tree
GO := go

.PHONY: help fmt vet tidy build install run test clean

help:
	@echo "Makefile targets:"
	@echo "  make build       - build binary into ./bin/$(BINARY)"
	@echo "  make install     - go install the CLI into GOPATH/bin"
	@echo "  make run ARGS=.. - run the CLI (pass ARGS, e.g. ARGS='--depth 1 .')"
	@echo "  make fmt         - go fmt ./..."
	@echo "  make vet         - go vet ./..."
	@echo "  make tidy        - go mod tidy"
	@echo "  make test        - go test ./..."
	@echo "  make clean       - remove ./bin"

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

tidy:
	$(GO) mod tidy

build:
	@echo "Building $(BINARY)"
	@mkdir -p bin
	$(GO) build -o bin/$(BINARY) ./cmd/tree

install:
	$(GO) install ./cmd/tree

run:
	$(GO) run ./cmd/tree $(ARGS)

test:
	$(GO) test ./...

clean:
	rm -rf bin
