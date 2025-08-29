.PHONY: setup submodules build-whisper build build-debug run clean help

BINARY_NAME=focus-helper
DEBUG_BINARY_NAME=focus-helper-debug
WHISPER_DIR := $(abspath ./whisper.cpp)
CGO_CFLAGS := "-I$(WHISPER_DIR)/include -I$(WHISPER_DIR)/ggml/include"
CGO_LDFLAGS := "-L$(WHISPER_DIR)/build/src -L$(WHISPER_DIR)/build/ggml/src -lwhisper -lggml"
GOCMD=go

.DEFAULT_GOAL := help

setup: submodules build-whisper
	@echo "✅ Setup ready! Now can compile"

build: build-whisper
	@echo "Building...: $(BINARY_NAME)... ---"
	CGO_ENABLED=1 CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) $(GOCMD) build -ldflags="-w -s" -o $(BINARY_NAME) main.go
	@echo "✅ Bin '$(BINARY_NAME)' ready"

build-debug: build-whisper
	@echo "Building debug...: $(DEBUG_BINARY_NAME)... ---"
	CGO_ENABLED=1 CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) $(GOCMD) build -gcflags="all=-N -l" -o $(DEBUG_BINARY_NAME) main.go
	@echo "✅ Bin '$(DEBUG_BINARY_NAME)' ready"

run: build-whisper
	@echo "--- running in development mode... ---"
	CGO_ENABLED=1 CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) $(GOCMD) run main.go -debug

clean:
	@echo "clear cache"
	rm -f $(BINARY_NAME) $(DEBUG_BINARY_NAME)
	rm -rf ./whisper.cpp/build
	@echo "✅ clear done."

submodules:
	@echo "updating git submodules"
	git submodule update --init --recursive

build-whisper:
	@echo "Compilling whisper and C++ dependencies"
	cmake ./whisper.cpp -B ./whisper.cpp/build -DBUILD_SHARED_LIBS=OFF
	cmake --build ./whisper.cpp/build --config Release -j$(nproc)

help:
	@echo "Available commands:"
	@echo "  make setup          - Set up the project for the first time (initializes submodules and builds C++)."
	@echo "  make build          - Compile the production binary."
	@echo "  make run            - Compile and execute the program."
	@echo "  make build-debug    - Create a debug binary for use with VS Code."
	@echo "  make clean          - Remove all compiled files and artifacts."