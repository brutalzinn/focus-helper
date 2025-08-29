.PHONY: setup submodules build-whisper build build-debug run clean help

BINARY_NAME=focushelper
DEBUG_BINARY_NAME=focushelper-debug

PROFILES_JSON := ./profiles.json

CONFIG_DIR := $(HOME)/.config/$(BINARY_NAME)

LANGS_DIR := langs
ASSETS_DIR := assets
VOICES_DIR := voices

DEST_LANGS_DIR := $(CONFIG_DIR)/langs
DEST_ASSETS_DIR := $(CONFIG_DIR)/assets
DEST_PROFILES_JSON := $(CONFIG_DIR)/profiles.json
DEST_VOICES_DIR := $(CONFIG_DIR)/voices


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

move-binary:
	@echo "Moving the binary to /usr/local/bin..."
	@sudo cp $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@sudo chmod +x /usr/local/bin/$(BINARY_NAME)

copy-langs:
	@echo "copying $(LANGS_DIR) to: $(DEST_LANGS_DIR)" 
	@mkdir -p $(DEST_LANGS_DIR)
	@cp -r $(LANGS_DIR)/. $(DEST_LANGS_DIR)

copy-voices:
	@mkdir -p $(DEST_VOICES_DIR)
	@echo "copying $(VOICES_DIR) to: $(DEST_VOICES_DIR)" 
	@cp -r $(VOICES_DIR)/. $(DEST_VOICES_DIR)

copy-assets:
	@mkdir -p $(DEST_ASSETS_DIR)
	@echo "copying $(ASSETS_DIR) to: $(DEST_ASSETS_DIR)" 
	@cp -r $(ASSETS_DIR)/. $(DEST_ASSETS_DIR)

copy-profiles:
	@echo "copying profiles.json to: $(DEST_PROFILES_JSON)" 
	@cp $(PROFILES_JSON) $(DEST_PROFILES_JSON)

move: copy-langs copy-voices copy-assets copy-profiles move-binary

help:
	@echo "Available commands:"
	@echo "  make setup          - Set up the project for the first time (initializes submodules and builds C++)."
	@echo "  make build          - Compile the production binary."
	@echo "  make run            - Compile and execute the program."
	@echo "  make build-debug    - Create a debug binary for use with VS Code."
	@echo "  make clean          - Remove all compiled files and artifacts."
	@echo "  make move         	 - Move all artifacts to your .config folder and set focushelper global by cmd tools"