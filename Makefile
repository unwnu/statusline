APP := statusline
VER ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
BUILD ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X main.version=$(VER) -X main.build=$(BUILD)

.PHONY: build lint test clean xwin xlinux xmac release

build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o .bin/$(APP) ./cmd/statusline

lint:
	golangci-lint run

test:
	go test ./...

clean:
	rm -rf .bin dist

# cross-compilation
xwin:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o .bin/$(APP).exe ./cmd/statusline
xlinux:
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o .bin/$(APP)     ./cmd/statusline
xmac:
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o .bin/$(APP)     ./cmd/statusline

# release
release:
	goreleaser release --clean