.PHONY: setup build run test clean help dev-setup dev-run dev-test dev-clean install-deps check-deps

BINARY_NAME=focushelper
DEBUG_BINARY_NAME=focushelper-debug

CONFIG_DIR := $(HOME)/.config/$(BINARY_NAME)
WHISPER_DIR := $(abspath ./whisper.cpp)
CGO_CFLAGS := "-I$(WHISPER_DIR)/include -I$(WHISPER_DIR)/ggml/include"
CGO_LDFLAGS := "-L$(WHISPER_DIR)/build/src -L$(WHISPER_DIR)/build/ggml/src -lwhisper -lggml"
GOCMD=go

.DEFAULT_GOAL := help

setup: submodules build-whisper
	@echo "✅ Setup ready! Now can compile"

build: build-whisper
	@echo "Building $(BINARY_NAME)..."
	CGO_ENABLED=1 CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) $(GOCMD) build -ldflags="-w -s" -o $(BINARY_NAME) src/cmd/focus-helper/main.go
	@echo "✅ Binary ready"

build-debug: build-whisper
	@echo "Building debug $(DEBUG_BINARY_NAME)..."
	CGO_ENABLED=1 CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) $(GOCMD) build -gcflags="all=-N -l" -o $(DEBUG_BINARY_NAME) src/cmd/focus-helper/main.go
	@echo "✅ Debug binary ready"

run: build-whisper
	@echo "Running in development mode..."
	CGO_ENABLED=1 CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) $(GOCMD) run src/cmd/focus-helper/main.go -debug

test:
	@echo "Running tests..."
	@$(GOCMD) test ./src/... -v
	@echo "✅ Tests completed"

test-coverage:
	@echo "Running tests with coverage..."
	@$(GOCMD) test ./src/... -coverprofile=coverage.out
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: coverage.html"

test-package:
	@if [ -z "$(PACKAGE)" ]; then echo "Usage: make test-package PACKAGE=src/pkg/actions"; exit 1; fi
	@$(GOCMD) test $(PACKAGE) -v

install: build
	@echo "Installing $(BINARY_NAME)..."
	@sudo cp $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@mkdir -p $(CONFIG_DIR)
	@cp -r langs $(CONFIG_DIR)/
	@cp -r voices $(CONFIG_DIR)/
	@mkdir -p $(CONFIG_DIR)/assets
	@cp -r views $(CONFIG_DIR)/
	@cp profiles.json $(CONFIG_DIR)/
	@echo "✅ Installed successfully"

clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME) $(DEBUG_BINARY_NAME) coverage.out coverage.html
	@rm -rf ./whisper.cpp/build
	@echo "✅ Clean completed"

submodules:
	@echo "Updating git submodules..."
	@git submodule update --init --recursive

build-whisper:
	@echo "Building whisper dependencies..."
	@cmake ./whisper.cpp -B ./whisper.cpp/build -DBUILD_SHARED_LIBS=OFF
	@cmake --build ./whisper.cpp/build --config Release -j$(nproc)

# Development Commands
dev-setup:
	@echo "🚀 Setting up development environment..."
	@chmod +x scripts/setup-dev.sh
	@./scripts/setup-dev.sh

dev-run:
	@echo "🏃 Running in development mode..."
	@make run

dev-test:
	@echo "🧪 Running development tests..."
	@make test-coverage
	@echo "📊 Coverage report: coverage.html"

dev-clean:
	@echo "🧹 Cleaning development artifacts..."
	@make clean
	@rm -f profiles-dev.json
	@rm -f focus_helper*.log
	@rm -f focus_helper*.db
	@echo "✅ Development cleanup completed"

# Dependency Management
install-deps:
	@echo "📦 Installing system dependencies..."
	@chmod +x scripts/setup-dev.sh
	@./scripts/setup-dev.sh

check-deps:
	@echo "🔍 Checking system dependencies..."
	@go version
	@cmake --version
	@make --version
	@git --version
	@echo "✅ Dependency check completed"

# Voice Management
download-voices:
	@echo "🎤 Downloading voice models..."
	@chmod +x scripts/download-voices.sh
	@./scripts/download-voices.sh

# Documentation
docs-serve:
	@echo "📚 Starting documentation server..."
	@cd docs && hugo server -D

docs-build:
	@echo "📚 Building documentation..."
	@cd docs && hugo --minify

# Docker Commands
docker-build:
	@echo "🐳 Building Docker image..."
	@docker build -t focus-helper .

docker-run:
	@echo "🐳 Running Docker container..."
	@docker run -it --rm --device /dev/snd focus-helper

# Quick Development
quick-start: check-deps dev-setup dev-run

help:
	@echo "Focus Helper - Development Commands"
	@echo "==================================="
	@echo ""
	@echo "🚀 Quick Start"
	@echo "=============="
	@echo "  make quick-start    - Complete setup and run (first time)"
	@echo "  make dev-setup      - Setup development environment"
	@echo "  make dev-run        - Run in development mode"
	@echo "  make dev-test       - Run tests with coverage"
	@echo "  make dev-clean      - Clean development artifacts"
	@echo ""
	@echo "🔧 Build Commands"
	@echo "================"
	@echo "  make setup          - Initial setup (submodules + whisper build)"
	@echo "  make build          - Build production binary"
	@echo "  make build-debug    - Build debug binary"
	@echo "  make run            - Run in development mode"
	@echo ""
	@echo "🧪 Testing Commands"
	@echo "=================="
	@echo "  make test           - Run all tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make test-package   - Test specific package (PACKAGE=src/pkg/name)"
	@echo ""
	@echo "📦 Dependencies"
	@echo "==============="
	@echo "  make install-deps   - Install system dependencies"
	@echo "  make check-deps     - Check system requirements"
	@echo "  make download-voices - Download voice models"
	@echo ""
	@echo "📚 Documentation"
	@echo "==============="
	@echo "  make docs-serve     - Start documentation server"
	@echo "  make docs-build     - Build documentation"
	@echo ""
	@echo "🐳 Docker"
	@echo "========"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-run     - Run Docker container"
	@echo ""
	@echo "🔧 Installation & Cleanup"
	@echo "========================="
	@echo "  make install        - Install binary and config files"
	@echo "  make clean          - Clean build artifacts"
	@echo ""
	@echo "📖 Examples:"
	@echo "  make test-package PACKAGE=src/pkg/actions"
	@echo "  make test-coverage && open coverage.html"
	@echo "  make quick-start    # First time setup"
	@echo "  make dev-run        # Daily development"