APP_NAME := focushelper
BIN_DIR := $(HOME)/.config/$(APP_NAME)
CONFIG_DIR := $(HOME)/.config/$(APP_NAME)
LANGS_DIR := langs
ASSETS_DIR := assets
PROFILES_JSON := profiles.json

DEST_LANGS_DIR := $(CONFIG_DIR)/langs
DEST_ASSETS_DIR := $(CONFIG_DIR)/assets
DEST_PROFILES_JSON := $(CONFIG_DIR)/profiles.json

GO := go
GO_BUILD := $(GO) build
GO_INSTALL := $(GO) install
BIN_PATH := $(BIN_DIR)/$(APP_NAME)

install: build copy-langs copy-assets copy-profiles move-binary

build:
	@echo "Building the Go binary..."
	$(GO_BUILD) -o $(APP_NAME) cmd/focus-helper/main.go

copy-langs:
	@echo "Copying langs directory..."
	@mkdir -p $(DEST_LANGS_DIR)
	@cp -r $(LANGS_DIR)/. $(DEST_LANGS_DIR)

copy-assets:
	@echo "Copying assets directory..."
	@mkdir -p $(DEST_ASSETS_DIR)
	@cp -r $(ASSETS_DIR)/. $(DEST_ASSETS_DIR)

copy-profiles:
	@echo "Copying profiles.json..."
	@cp $(PROFILES_JSON) $(DEST_PROFILES_JSON)

move-binary:
	@echo "Moving the binary to /usr/local/bin..."
	@sudo mv $(APP_NAME) /usr/local/bin/$(APP_NAME)
	@sudo chmod +x /usr/local/bin/$(APP_NAME)

clean:
	@echo "Cleaning up..."
	@rm -rf $(CONFIG_DIR) $(APP_NAME)

help:
	@echo "Available targets:"
	@echo "  install       - Build and install the binary and copy the config files"
	@echo "  clean         - Clean up the build files and config directories"
	@echo "  build         - Build the Go binary"
	@echo "  copy-langs    - Copy the langs directory to the config directory"
	@echo "  copy-assets   - Copy the assets directory to the config directory"
	@echo "  copy-profiles - Copy the profiles.json to the config directory"
	@echo "  move-binary   - Move the binary to a globally available directory"
