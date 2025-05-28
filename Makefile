.PHONY: build test clean lint vet fmt help release install packages publish build-ui dev-ui deps build-all clean-ui build-frontend

BINARY_NAME=insta
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Define supported OS/ARCH combinations
PLATFORMS=linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

help:
	@echo "insta-infra - A tool for running data infrastructure services"
	@echo ""
	@echo "Usage:"
	@echo "  make build       Build CLI binary"
	@echo "  make build-ui    Build Web UI binary (production)"
	@echo "  make build-ui-bundled Build Web UI with bundled CLI (recommended for distribution)"
	@echo "  make build-frontend Build frontend assets only (for embed)"
	@echo "  make build-all   Build both CLI and Web UI"
	@echo "  make dev-ui      Start Web UI in development mode"
	@echo "  make deps        Install all dependencies (Go, Node.js, Wails)"
	@echo "  make test        Run all tests (Go + Frontend)"
	@echo "  make test-go     Run Go tests only"
	@echo "  make test-go-coverage Run Go tests with coverage"
	@echo "  make test-ui     Run frontend tests only"
	@echo "  make test-ui-coverage Run frontend tests with coverage"
	@echo "  make clean       Clean all build artifacts"
	@echo "  make clean-ui    Clean Web UI build artifacts only"
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

build-ui:
	@echo "Preparing UI resources..."
	@chmod +x scripts/prepare-ui-resources.sh
	@./scripts/prepare-ui-resources.sh
	@echo "Building Wails UI application (production)..."
	cd cmd/instaui && wails build
	@echo "Wails UI application built. Binary available in cmd/instaui/build/bin/"

build-ui-bundled: build
	@echo "Preparing UI resources..."
	@chmod +x scripts/prepare-ui-resources.sh
	@./scripts/prepare-ui-resources.sh
	@echo "Building Wails UI application with bundled CLI (production)..."
	cd cmd/instaui && wails build -clean
	@echo "Bundling CLI binary into app..."
	./scripts/bundle-cli.sh
	@echo "Wails UI application with bundled CLI built. Binary available in cmd/instaui/build/bin/"

dev-ui:
	@echo "Starting Wails UI application (development mode)..."
	cd cmd/instaui && wails dev

test: test-go test-ui

test-go:
	@echo "Preparing UI resources for tests..."
	@chmod +x scripts/prepare-ui-resources.sh
	@./scripts/prepare-ui-resources.sh
	@echo "Running Go tests..."
	go test -v ./...

test-go-coverage:
	@echo "Running Go tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-ui:
	@echo "Running frontend tests..."
	@if [ -d "cmd/instaui/frontend/node_modules" ]; then \
		cd cmd/instaui/frontend && npm test; \
	else \
		echo "Frontend dependencies not installed. Run 'make deps' first."; \
		exit 1; \
	fi

test-ui-coverage:
	@echo "Running frontend tests with coverage..."
	@if [ -d "cmd/instaui/frontend/node_modules" ]; then \
		cd cmd/instaui/frontend && npm run test:coverage; \
	else \
		echo "Frontend dependencies not installed. Run 'make deps' first."; \
		exit 1; \
	fi

clean: clean-ui
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*.tar.gz
	rm -f $(BINARY_NAME)-*.zip
	rm -f $(BINARY_NAME)-*-*
	rm -rf build/
	rm -rf release/

lint:
	@which golint > /dev/null || go install golang.org/x/lint/golint@latest
	golint ./...

vet: build-frontend
	@echo "Preparing UI resources for vet..."
	@chmod +x scripts/prepare-ui-resources.sh
	@./scripts/prepare-ui-resources.sh
	go vet ./...

build-frontend:
	@echo "Building frontend for embed..."
	@if [ -d "cmd/instaui/frontend/node_modules" ]; then \
		cd cmd/instaui/frontend && npm run build; \
	else \
		echo "Frontend dependencies not installed. Run 'make deps' first."; \
		exit 1; \
	fi

fmt:
	@echo "Running go fmt..."
	go fmt ./...
	@echo "Checking for npm..."
	@if ! command -v npm > /dev/null; then \
		echo "npm is not installed. Skipping YAML formatting."; \
	else \
		echo "Installing ESLint and YAML plugin (if not already installed or for updates)..."; \
		npm install eslint @eslint/js eslint-plugin-yml; \
		echo "Formatting cmd/insta/resources/docker-compose.yaml..."; \
		npx eslint cmd/insta/resources/docker-compose.yaml --fix || true; \
		echo "Formatting cmd/insta/resources/docker-compose-persist.yaml..."; \
		npx eslint cmd/insta/resources/docker-compose-persist.yaml --fix || true; \
		echo "YAML formatting complete."; \
	fi

packages:
	@chmod +x scripts/packaging.sh
	@VERSION=$(VERSION) BUILD_TIME=$(BUILD_TIME) RELEASE=true BUILD_PACKAGES=true ./scripts/packaging.sh

release: clean
	@mkdir -p release
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*} GOARCH=$${platform##*/} CGO_ENABLED=0 VERSION=$(VERSION) BUILD_TIME=$(BUILD_TIME) RELEASE=true ./scripts/build.sh; \
	done
	@chmod +x scripts/packaging.sh
	@VERSION=$(VERSION) BUILD_TIME=$(BUILD_TIME) RELEASE=true BUILD_PACKAGES=false ./scripts/packaging.sh

publish: clean
	@chmod +x scripts/packaging.sh
	@VERSION=$(VERSION) BUILD_TIME=$(BUILD_TIME) RELEASE=true PUBLISH=true ./scripts/packaging.sh

install: build
	mv $(BINARY_NAME) $(GOPATH)/bin/

deps:
	@echo "Installing Go dependencies..."
	@go mod download
	@echo "Installing Wails CLI..."
	@go install github.com/wailsapp/wails/v2/cmd/wails@latest
	@echo "Installing Web UI dependencies..."
	@cd cmd/instaui/frontend && npm install
	@echo "All dependencies installed successfully!"

build-all: build build-ui
	@echo "Both CLI and Web UI built successfully!"

clean-ui:
	@echo "Cleaning Web UI build artifacts..."
	@rm -rf cmd/instaui/build
	@rm -rf cmd/instaui/frontend/dist

clean-all: clean clean-ui 