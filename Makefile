.PHONY: help install dev build prod clean test docker-build docker-run css watch db-reset docker-down docs-fmt docker-logs docker-login docker-push

SHELL := /bin/bash
	export
	include docker/.env.local

# Get the version from git, or default to dev
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0-dev")

GH_USER ?= $(GITHUB_USERNAME)
REGISTRY := ghcr.io
IMAGE_NAME := $(GH_USER)/kindredcard

# Default target - show help
help:
	@echo "KindredCard - Available Commands"
	@echo ""
	@echo "Setup:"
	@echo "  make install      - Install all dependencies (Go + Node)"
	@echo ""
	@echo "Development:"
	@echo "  make dev          - Run development servers (CSS watch + Go server)"
	@echo "  make watch        - Watch CSS changes only"
	@echo "  make css          - Build CSS once"
	@echo ""
	@echo "Database:"
	@echo "  make db-reset     - Reset database with fresh schema"
	@echo ""
	@echo "Production:"
	@echo "  make build        - Build production binary"
	@echo "  make prod         - Build optimized production binary"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run   - Run Docker container"
	@echo ""
	@echo "Cleanup:"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make test         - Run tests"

# Install all dependencies
install:
	@echo "📦 Installing Go dependencies..."
	go mod download
	@echo "📦 Installing Node.js dependencies..."
	npm install
	@echo "✅ Dependencies installed!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Set environment variables (see .env.example)"
	@echo "  2. Run: make db-reset"
	@echo "  3. Run: make dev"

docker-down:
	@docker compose -f 'docker/docker-compose.yml' down

docker-logs:
	@docker compose -f 'docker/docker-compose.yml' logs -f

docker-restart:
	make docker-down
	make docker-run

docker-pg-up:
	docker compose -f 'docker/docker-compose.yml' up -d --build 'postgres'

# Development mode - watch CSS and run Go server
dev:
	@echo "🚀 Starting KindredCard development servers..."
	make docs-fmt
	make css
	@echo "   CSS will rebuild automatically on changes"
	@echo "   Press Ctrl+C to stop"
	@echo ""
	@npm run watch:css & \
	go run cmd/kindredcard/main.go

docs-fmt: ## Format swagger comments
	swag fmt -g cmd/kindredcard/main.go
	swag init -g cmd/kindredcard/main.go

# Watch CSS changes only
watch:
	npm run watch:css

# Build CSS once
css:
	@echo "🎨 Building CSS..."
	npm run build:css
	@echo "✅ CSS built!"

# Reset database
db-reset:
	@echo "🗑️  Resetting database..."
	make docker-down
	docker volume rm kindredcard_postgres_data
	make docker-pg-up
	@echo "✅ Database reset complete!"

# Build for production with optimized CSS
build: css
	@echo "🔨 Building KindredCard $(VERSION)..."
	go build -ldflags="-X 'main.ReleaseVersion=$(VERSION)'" -o kindredcard cmd/kindredcard/main.go
	@echo "✅ Build complete: ./kindredcard"

# Production build with minified CSS and optimized binary
prod:
	@echo "🔨 Building production KindredCard $(VERSION)..."
	npm run build:css:prod
	go build -ldflags="-s -w -X 'main.ReleaseVersion=$(VERSION)'" -o kindredcard cmd/kindredcard/main.go
	@echo "✅ Production build complete!"
	@echo "   Binary: ./kindredcard"
	@ls -lh kindredcard

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -f kindredcard
	rm -f web/static/css/output.css
	rm -f web/static/css/output.css.map
	@echo "✅ Clean complete!"

# Run tests
test:
	@echo "🧪 Running tests..."
	go test ./... -v

docker-build:
	@echo "🐳 Building Docker image $(VERSION)..."
	docker build \
		--build-arg VERSION=$(VERSION) \
		-t kindredcard:latest \
		-f docker/Dockerfile .
	@echo "✅ Docker image built: kindredcard:latest"

# Docker run
docker-run:
	@echo "🐳 Running Docker container..."
	make docker-build
	docker compose -f 'docker/docker-compose.yml' up -d
	@echo "✅ Container running at http://localhost:8080"
	@echo "   View logs: docker logs -f kindredcard"
	@echo "   Stop: docker stop kindredcard"
	@echo "   Remove: docker rm kindredcard"

# Quick restart (for development)
restart:
	@echo "🔄 Restarting..."
	@pkill -f "kindredcard" || true
	@make dev

# Authenticate using the GITHUB_TOKEN from your .bashrc
docker-login:
	@echo "🔑 Logging into GHCR..."
	@echo "$(GITHUB_TOKEN)" | docker login $(REGISTRY) -u $(GH_USER) --password-stdin

# Build, tag, and push both 'latest' and the specific version
docker-push: docker-login docker-build
	@echo "🚀 Pushing images to $(REGISTRY)..."
	docker tag kindredcard:latest $(REGISTRY)/$(IMAGE_NAME):latest
	docker tag kindredcard:latest $(REGISTRY)/$(IMAGE_NAME):$(VERSION)
	docker push $(REGISTRY)/$(IMAGE_NAME):latest
	docker push $(REGISTRY)/$(IMAGE_NAME):$(VERSION)
	@echo "✅ Images pushed successfully!"