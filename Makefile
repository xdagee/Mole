# Makefile for Mole

.PHONY: all build build-gui clean check format test test-go verify release release-amd64 release-arm64 release-windows-amd64 mod-download

# Output directory
BIN_DIR := bin

# Go toolchain
GO ?= go
GO_DOWNLOAD_RETRIES ?= 3

# Binaries
MOLE := mole
ANALYZE := analyze
STATUS := status

# Source directories
MOLE_SRC := ./cmd/mole
ANALYZE_SRC := ./cmd/analyze
STATUS_SRC := ./cmd/status

# Build flags
LDFLAGS := -s -w

all: build build-gui

# Download modules with retries to mitigate transient proxy/network EOF errors.
mod-download:
	@attempt=1; \
	while [ $$attempt -le $(GO_DOWNLOAD_RETRIES) ]; do \
		echo "Downloading Go modules ($$attempt/$(GO_DOWNLOAD_RETRIES))..."; \
		if $(GO) mod download; then \
			exit 0; \
		fi; \
		sleep $$((attempt * 2)); \
		attempt=$$((attempt + 1)); \
	done; \
	echo "Go module download failed after $(GO_DOWNLOAD_RETRIES) attempts"; \
	exit 1

# Local build (current architecture)
build: mod-download
	@echo "Building for local architecture..."
	$(GO) build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(MOLE)-go $(MOLE_SRC)
	$(GO) build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(ANALYZE)-go $(ANALYZE_SRC)
	$(GO) build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(STATUS)-go $(STATUS_SRC)

build-gui: mod-download
	@echo "Building GUI (Windows only)..."
	cd cmd/gui && wails build -platform windows/amd64
	cp cmd/gui/build/bin/gui.exe $(BIN_DIR)/mole-gui-windows-amd64.exe || true

check:
	./scripts/check.sh --no-format

format:
	./scripts/check.sh --format

test:
	MOLE_TEST_NO_AUTH=1 ./scripts/test.sh

test-go:
	$(GO) test ./...

verify: check test-go

# Release build targets (run on native architectures for CGO support)
release-amd64: mod-download
	@echo "Building release binaries (amd64)..."
	GOOS=darwin GOARCH=amd64 $(GO) build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(MOLE)-darwin-amd64 $(MOLE_SRC)
	GOOS=darwin GOARCH=amd64 $(GO) build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(ANALYZE)-darwin-amd64 $(ANALYZE_SRC)
	GOOS=darwin GOARCH=amd64 $(GO) build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(STATUS)-darwin-amd64 $(STATUS_SRC)

release-arm64: mod-download
	@echo "Building release binaries (arm64)..."
	GOOS=darwin GOARCH=arm64 $(GO) build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(MOLE)-darwin-arm64 $(MOLE_SRC)
	GOOS=darwin GOARCH=arm64 $(GO) build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(ANALYZE)-darwin-arm64 $(ANALYZE_SRC)
	GOOS=darwin GOARCH=arm64 $(GO) build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(STATUS)-darwin-arm64 $(STATUS_SRC)

release-windows-amd64: mod-download
	@echo "Building release binaries (windows amd64)..."
	GOOS=windows GOARCH=amd64 $(GO) build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(MOLE)-windows-amd64.exe $(MOLE_SRC)
	GOOS=windows GOARCH=amd64 $(GO) build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(ANALYZE)-windows-amd64.exe $(ANALYZE_SRC)
	GOOS=windows GOARCH=amd64 $(GO) build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(STATUS)-windows-amd64.exe $(STATUS_SRC)
	@echo "Building GUI release binary (windows amd64)..."
	cd cmd/gui && wails build -platform windows/amd64 -s -w
	cp cmd/gui/build/bin/gui.exe $(BIN_DIR)/mole-gui-windows-amd64.exe || true

clean:
	@echo "Cleaning binaries..."
	rm -f $(BIN_DIR)/$(MOLE)-* $(BIN_DIR)/$(ANALYZE)-* $(BIN_DIR)/$(STATUS)-* $(BIN_DIR)/$(MOLE)-go $(BIN_DIR)/$(ANALYZE)-go $(BIN_DIR)/$(STATUS)-go
