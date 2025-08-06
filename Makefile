.PHONY: build test clean lint vet fmt help release install packages publish deps build-all build-frontend build-web dev-web clean-web clean-test

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
	@echo "  make build-web   Build browser-based Web UI binary (recommended)"
	@echo "  make build-frontend Build frontend assets only (for embed)"
	@echo "  make build-all   Build both CLI and Web UI"
	@echo "  make build-demo  Build demo UI"
	@echo "  make dev-web     Start browser-based Web UI in development mode"
	@echo "  make deps        Install all dependencies (Go, Node.js)"
	@echo "  make test        Run all tests (Go + Frontend)"
	@echo "  make test-go     Run Go tests only"
	@echo "  make test-go-coverage Run Go tests with coverage"
	@echo "  make test-ui     Run frontend tests only"
	@echo "  make test-ui-coverage Run frontend tests with coverage"
	@echo "  make clean-test  Clean test cache and temporary files"
	@echo "  make clean       Clean all build artifacts"
	@echo "  make clean-web   Clean browser-based Web UI build artifacts only"
	@echo "  make lint        Run linter"
	@echo "  make vet         Run go vet"
	@echo "  make fmt         Run go fmt"
	@echo "  make release     Create release archive"
	@echo "  make install     Install binary to GOPATH/bin"
	@echo "  make help        Show this help message"

build:
	@chmod +x scripts/build.sh
	@VERSION=$(VERSION) BUILD_TIME=$(BUILD_TIME) RELEASE=false ./scripts/build.sh



test: clean-test build-web test-go test-ui

test-go:
	@echo "Running Go tests..."
	go clean -testcache
	go test -v -count=1 ./...

test-go-coverage:
	@echo "Running Go tests with coverage..."
	go clean -testcache
	go test -v -count=1 -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-ui:
	@echo "Running frontend tests..."
	@if [ -d "cmd/insta/frontend/node_modules" ]; then \
		cd cmd/insta/frontend && npm test; \
	else \
		echo "Frontend dependencies not installed. Run 'make deps' first."; \
		exit 1; \
	fi

clean-test:
	@echo "Cleaning test cache and temporary files..."
	go clean -testcache
	go clean -cache
	rm -f coverage.out coverage.html
	rm -rf /tmp/app-test-*

test-ui-coverage:
	@echo "Running frontend tests with coverage..."
	@if [ -d "cmd/insta/frontend/node_modules" ]; then \
		cd cmd/insta/frontend && npm run test:coverage; \
	else \
		echo "Frontend dependencies not installed. Run 'make deps' first."; \
		exit 1; \
	fi

clean: clean-web
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

build-frontend:
	@echo "Building frontend for embed..."
	@if [ -d "cmd/insta/frontend/node_modules" ]; then \
		cd cmd/insta/frontend && npm run build; \
	else \
		echo "Frontend dependencies not installed. Run 'make deps' first."; \
		exit 1; \
	fi

build-demo:
	@echo "Building demo UI..."
	@cd cmd/insta/frontend && npm run build:demo

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

release: clean build-web
	@chmod +x scripts/packaging.sh
	@VERSION=$(VERSION) BUILD_TIME=$(BUILD_TIME) ./scripts/packaging.sh

install: build
	mv $(BINARY_NAME) $(GOPATH)/bin/

deps:
	@echo "Installing Go dependencies..."
	@go mod download
	@echo "Installing Web UI dependencies..."
	@cd cmd/insta/frontend && npm install
	@echo "All dependencies installed successfully!"

build-web: clean-web
	@echo "Building browser-based Web UI..."
	@chmod +x scripts/build-web.sh
	@VERSION=$(VERSION) BUILD_TIME=$(BUILD_TIME) ./scripts/build-web.sh

dev-web:
	@echo "Starting browser-based Web UI in development mode..."
	@echo "Building frontend..."
	@cd cmd/insta/frontend && npm run build
	@echo "Copying frontend to embed location..."
	@cp -r cmd/insta/frontend/dist/* cmd/insta/ui/dist/
	@echo "Starting web server..."
	@cd cmd/insta && go run . --ui

clean-web:
	@echo "Cleaning browser-based Web UI build artifacts..."
	@rm -rf cmd/insta/frontend/dist
	@rm -rf cmd/insta/ui/dist/*

build-all: build build-web
	@echo "Both CLI and browser-based Web UI built successfully!"

clean-all: clean clean-web 