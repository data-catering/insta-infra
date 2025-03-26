.PHONY: build test clean lint vet fmt help release install packages publish

BINARY_NAME=insta
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

help:
	@echo "insta-infra - A tool for running data infrastructure services"
	@echo ""
	@echo "Usage:"
	@echo "  make build       Build binary"
	@echo "  make test        Run tests"
	@echo "  make clean       Clean build artifacts"
	@echo "  make lint        Run linter"
	@echo "  make vet         Run go vet"
	@echo "  make fmt         Run go fmt"
	@echo "  make release     Create release archive and packages"
	@echo "  make packages    Build system packages (Debian, RPM, Arch)"
	@echo "  make publish     Build and publish packages (requires environment variables)"
	@echo "  make install     Install binary to GOPATH/bin"
	@echo "  make help        Show this help message"

build:
	@chmod +x scripts/build.sh
	@VERSION=$(VERSION) BUILD_TIME=$(BUILD_TIME) RELEASE=false ./scripts/build.sh

test:
	go test -v ./...

clean:
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*.tar.gz
	rm -f $(BINARY_NAME)-*.zip
	rm -f $(BINARY_NAME)-*-*
	rm -rf build/
	rm -rf release/

lint:
	@which golint > /dev/null || go install golang.org/x/lint/golint@latest
	golint ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

packages:
	@chmod +x scripts/packaging.sh
	@VERSION=$(VERSION) BUILD_TIME=$(BUILD_TIME) RELEASE=true ./scripts/packaging.sh

release: clean
	@chmod +x scripts/packaging.sh
	@VERSION=$(VERSION) BUILD_TIME=$(BUILD_TIME) RELEASE=true ./scripts/packaging.sh

publish: clean
	@chmod +x scripts/packaging.sh
	@VERSION=$(VERSION) BUILD_TIME=$(BUILD_TIME) RELEASE=true PUBLISH=true ./scripts/packaging.sh

install: build
	mv $(BINARY_NAME) $(GOPATH)/bin/ 