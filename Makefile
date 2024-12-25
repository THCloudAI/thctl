.PHONY: build test clean

VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

LDFLAGS := -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)

build:
	go build -ldflags "$(LDFLAGS)" -o bin/thctl ./cmd/thctl

test:
	go test -v ./...

clean:
	rm -rf bin/
	go clean -cache

lint:
	golangci-lint run

deps:
	go mod tidy
	go mod verify

# Build for multiple platforms
build-all: clean
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/thctl-darwin-amd64 ./cmd/thctl
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/thctl-darwin-arm64 ./cmd/thctl
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/thctl-linux-amd64 ./cmd/thctl
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/thctl-windows-amd64.exe ./cmd/thctl

# Create release archives
release: build-all
	cd bin && tar czf thctl-darwin-amd64.tar.gz thctl-darwin-amd64
	cd bin && tar czf thctl-darwin-arm64.tar.gz thctl-darwin-arm64
	cd bin && tar czf thctl-linux-amd64.tar.gz thctl-linux-amd64
	cd bin && zip thctl-windows-amd64.zip thctl-windows-amd64.exe
