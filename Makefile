
# Citadel Makefile

SERVER := citadel-server
CLIENT := citadel-client

BIN := bin
GO := go
LDFLAGS := -s -w

.PHONY: all setup build linux windows \
        linux-amd64 linux-arm64 linux-arm \
        windows-amd64 windows-arm64 \
        clean help

# Download and verify Go dependencies
setup:
	$(GO) mod download
	$(GO) mod verify

# Build all supported platforms
all: linux windows

# Build for the current operating system and architecture
build: setup
	@mkdir -p $(BIN)
	$(GO) build -ldflags="$(LDFLAGS)" -o $(BIN)/$(SERVER) ./server
	$(GO) build -ldflags="$(LDFLAGS)" -o $(BIN)/$(CLIENT) ./client

# Build all supported Linux architectures
linux: linux-amd64 linux-arm64 linux-arm

linux-amd64: setup
	@mkdir -p $(BIN)/linux-amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -ldflags="$(LDFLAGS)" -o $(BIN)/linux-amd64/$(SERVER) ./server
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -ldflags="$(LDFLAGS)" -o $(BIN)/linux-amd64/$(CLIENT) ./client

linux-arm64: setup
	@mkdir -p $(BIN)/linux-arm64
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build -ldflags="$(LDFLAGS)" -o $(BIN)/linux-arm64/$(SERVER) ./server
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build -ldflags="$(LDFLAGS)" -o $(BIN)/linux-arm64/$(CLIENT) ./client

linux-arm: setup
	@mkdir -p $(BIN)/linux-arm
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 $(GO) build -ldflags="$(LDFLAGS)" -o $(BIN)/linux-arm/$(SERVER) ./server
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 $(GO) build -ldflags="$(LDFLAGS)" -o $(BIN)/linux-arm/$(CLIENT) ./client

# Build all supported Windows architectures
windows: windows-amd64 windows-arm64

windows-amd64: setup
	@mkdir -p $(BIN)/windows-amd64
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build -ldflags="$(LDFLAGS)" -o $(BIN)/windows-amd64/$(SERVER).exe ./server
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build -ldflags="$(LDFLAGS)" -o $(BIN)/windows-amd64/$(CLIENT).exe ./client

windows-arm64: setup
	@mkdir -p $(BIN)/windows-arm64
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 $(GO) build -ldflags="$(LDFLAGS)" -o $(BIN)/windows-arm64/$(SERVER).exe ./server
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 $(GO) build -ldflags="$(LDFLAGS)" -o $(BIN)/windows-arm64/$(CLIENT).exe ./client

# Remove locally built binaries
clean:
	rm -rf $(BIN)

# Display available targets
help:
	@echo "Citadel - Build System"
	@echo "  make setup    Download and verify dependencies"
	@echo "  make build    Build for the current platform"
	@echo "  make linux    Build all Linux architectures"
	@echo "  make windows  Build all Windows architectures"
	@echo "  make all      Build Linux and Windows"
	@echo "  make clean    Remove local binaries"
	@echo "  make help     Display this help"
