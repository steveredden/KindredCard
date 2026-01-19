.PHONY: help dev db-reset release docker-login setup-buildx clean

# Configuration
GH_USER ?= $(GITHUB_USERNAME)
REGISTRY := ghcr.io
IMAGE_NAME := $(REGISTRY)/$(GH_USER)/kindredcard
DOCKER_DIR := ./docker
DOCKER_COMPOSE := docker compose -f $(DOCKER_DIR)/docker-compose.yml

# Get version from git or default
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0-dev")

help:
	@echo "KindredCard Management"
	@echo "  make dev          - Run local dev server (Go + CSS watch)"
	@echo "  make db-reset     - Wipe database and start fresh" 
	@echo "  make release      - Prompt for version and push universal images to GHCR"
	@echo "  make clean        - Remove local build artifacts"

# --- Development ---

dev:
	@echo "🚀 Starting development servers..."
	@npm run build:css && (npm run watch:css & go run cmd/kindredcard/main.go)

db-reset:
	@echo "🗑️  Resetting database..."
	$(DOCKER_COMPOSE) down -v
	$(DOCKER_COMPOSE) up -d postgres 
	@echo "✅ Database is fresh and running."

# --- Production & Release ---

docker-login:
	@echo "🔑 Logging into GHCR..."
	@echo "$(GITHUB_TOKEN)" | docker login $(REGISTRY) -u $(GH_USER) --password-stdin

setup-buildx:
	@docker buildx create --name kindred-builder --use || true 
	@docker buildx inspect --bootstrap

release: setup-buildx docker-login
	@read -p "Enter version tag (e.g., v1.0.1): " REL_VER; \
	echo "🌎 Building and pushing universal images for $$REL_VER..."; \
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--provenance=false \
		--build-arg VERSION=$$REL_VER \
		-t $(IMAGE_NAME):$$REL_VER \
		-t $(IMAGE_NAME):latest \
		-f $(DOCKER_DIR)/Dockerfile \
		--push . ; \
	@echo "✅ Release $$REL_VER pushed successfully!"

clean:
	@echo "🧹 Cleaning..."
	rm -f kindredcard
	rm -f web/static/css/output.css*