BINARY := o2s
PKG := github.com/O2S-Y/o2s
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w \
  -X $(PKG)/internal/version.Version=$(VERSION) \
  -X $(PKG)/internal/version.Commit=$(COMMIT) \
  -X $(PKG)/internal/version.Date=$(DATE)

.PHONY: all build run tidy test fmt vet lint clean release-snapshot install

all: build

tidy:
	go mod tidy

build:
	go build -trimpath -ldflags "$(LDFLAGS)" -o bin/$(BINARY) ./cmd/o2s

install:
	go install -trimpath -ldflags "$(LDFLAGS)" ./cmd/o2s

run:
	go run ./cmd/o2s

test:
	go test ./...

fmt:
	gofmt -s -w .

vet:
	go vet ./...

clean:
	rm -rf bin dist

release-snapshot:
	goreleaser release --snapshot --clean
