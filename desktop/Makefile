APP_NAME := Microsoft Outlook
APP_SLUG := microsoft-outlook
ENTRYPOINT := ./cmd/outlook

BUILD_DIR := build
BIN_DIR := $(BUILD_DIR)/bin
DIST_DIR := $(BUILD_DIR)/dist
GOCACHE_DIR := $(CURDIR)/.gocache

ICON_FILE := assets/icon.icns
PLIST_FILE := packaging/macos/Info.plist

UNAME_S := $(shell uname -s)
NATIVE_GOARCH := $(shell go env GOARCH)

MACOS_ARM64_BIN := $(BIN_DIR)/darwin/arm64/$(APP_SLUG)
MACOS_ARM64_DIR := $(BIN_DIR)/darwin/arm64
MACOS_APP_BUNDLE := $(DIST_DIR)/$(APP_NAME).app
MACOS_APP_MACOS_DIR := $(MACOS_APP_BUNDLE)/Contents/MacOS
MACOS_APP_RESOURCES_DIR := $(MACOS_APP_BUNDLE)/Contents/Resources

LINUX_AMD64_BIN := $(BIN_DIR)/linux/amd64/$(APP_SLUG)
LINUX_ARM64_BIN := $(BIN_DIR)/linux/arm64/$(APP_SLUG)
LINUX_AMD64_DIR := $(BIN_DIR)/linux/amd64
LINUX_ARM64_DIR := $(BIN_DIR)/linux/arm64
LINUX_AMD64_ARCHIVE := $(DIST_DIR)/$(APP_SLUG)-linux-amd64.tar.gz
LINUX_ARM64_ARCHIVE := $(DIST_DIR)/$(APP_SLUG)-linux-arm64.tar.gz

ifeq ($(UNAME_S),Linux)
DEFAULT_BUILD_TARGET := build-linux
DEFAULT_BUNDLE_TARGET := bundle-linux
else
DEFAULT_BUILD_TARGET := build-macos
DEFAULT_BUNDLE_TARGET := bundle-macos
endif

.PHONY: all tidy run build bundle \
	build-macos bundle-macos \
	build-linux build-linux-amd64 build-linux-arm64 \
	bundle-linux bundle-linux-amd64 bundle-linux-arm64 \
	build-all bundle-all \
	clean

all: bundle

tidy:
	mkdir -p "$(GOCACHE_DIR)"
	GOCACHE="$(GOCACHE_DIR)" go mod tidy

run:
	mkdir -p "$(GOCACHE_DIR)"
	GOCACHE="$(GOCACHE_DIR)" go run $(ENTRYPOINT)

build: $(DEFAULT_BUILD_TARGET)

bundle: $(DEFAULT_BUNDLE_TARGET)

build-all: build-macos build-linux-amd64 build-linux-arm64

bundle-all: bundle-macos bundle-linux-amd64 bundle-linux-arm64

build-macos:
	mkdir -p "$(MACOS_ARM64_DIR)" "$(GOCACHE_DIR)"
	GOCACHE="$(GOCACHE_DIR)" GOOS=darwin GOARCH=arm64 go build -o "$(MACOS_ARM64_BIN)" $(ENTRYPOINT)

bundle-macos: build-macos
	rm -rf "$(MACOS_APP_BUNDLE)"
	mkdir -p "$(MACOS_APP_MACOS_DIR)" "$(MACOS_APP_RESOURCES_DIR)"
	cp "$(MACOS_ARM64_BIN)" "$(MACOS_APP_MACOS_DIR)/$(APP_NAME)"
	chmod +x "$(MACOS_APP_MACOS_DIR)/$(APP_NAME)"
	cp "$(ICON_FILE)" "$(MACOS_APP_RESOURCES_DIR)/icon.icns"
	cp "$(PLIST_FILE)" "$(MACOS_APP_BUNDLE)/Contents/Info.plist"

build-linux: build-linux-$(NATIVE_GOARCH)

build-linux-amd64:
	@if [ "$(UNAME_S)" != "Linux" ]; then echo "build-linux-amd64 requires a Linux host with GUI build dependencies."; exit 1; fi
	mkdir -p "$(LINUX_AMD64_DIR)" "$(GOCACHE_DIR)"
	GOCACHE="$(GOCACHE_DIR)" GOOS=linux GOARCH=amd64 go build -o "$(LINUX_AMD64_BIN)" $(ENTRYPOINT)

build-linux-arm64:
	@if [ "$(UNAME_S)" != "Linux" ]; then echo "build-linux-arm64 requires a Linux host with GUI build dependencies."; exit 1; fi
	mkdir -p "$(LINUX_ARM64_DIR)" "$(GOCACHE_DIR)"
	GOCACHE="$(GOCACHE_DIR)" GOOS=linux GOARCH=arm64 go build -o "$(LINUX_ARM64_BIN)" $(ENTRYPOINT)

bundle-linux: bundle-linux-$(NATIVE_GOARCH)

bundle-linux-amd64: build-linux-amd64
	mkdir -p "$(DIST_DIR)"
	tar -C "$(LINUX_AMD64_DIR)" -czf "$(LINUX_AMD64_ARCHIVE)" "$(APP_SLUG)"

bundle-linux-arm64: build-linux-arm64
	mkdir -p "$(DIST_DIR)"
	tar -C "$(LINUX_ARM64_DIR)" -czf "$(LINUX_ARM64_ARCHIVE)" "$(APP_SLUG)"

clean:
	rm -rf "$(BUILD_DIR)" "$(GOCACHE_DIR)"
