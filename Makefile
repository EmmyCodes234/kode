VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +%Y-%m-%d 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags="-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"
BINARY := bin/kode$(shell go env GOEXE)

.PHONY: all build test vet clean lint install release-checksum

all: build

build:
	go build $(LDFLAGS) -o $(BINARY) ./cmd/kode

test:
	go test ./...

vet:
	go vet ./...

lint: vet

install: build
	go install $(LDFLAGS) ./cmd/kode

clean:
	rm -rf bin/ dist/

# Cross-compile all platforms
release-checksum:
	@mkdir -p dist
	GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o dist/kode-linux-amd64 ./cmd/kode
	GOOS=linux   GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o dist/kode-linux-arm64 ./cmd/kode
	GOOS=darwin  GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o dist/kode-darwin-amd64 ./cmd/kode
	GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o dist/kode-darwin-arm64 ./cmd/kode
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o dist/kode-windows-amd64.exe ./cmd/kode
	cd dist && sha256sum * > checksums.txt
	@echo "Release assets in dist/"
