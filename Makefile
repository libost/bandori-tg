VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

BINARY_NAME=bandori-tg
BUILD_DIR=dist

LDFLAGS=-ldflags "-X github.com/libost/bandori-tg/version.Version=$(VERSION) -X github.com/libost/bandori-tg/version.Commit=$(COMMIT) -X github.com/libost/bandori-tg/version.Date=$(TIME)"

.PHONY: clean test vet tidy before-build build build-windows build-linux build-macos help

clean:
	rm -rf $(BUILD_DIR)

test:
	go test ./...

vet:
	go vet ./...

tidy:
	go mod tidy

before-build: tidy vet test

build: before-build
	mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) main.go

build-windows: before-build
	mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME).exe main.go

build-linux: before-build
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) main.go

build-macos: before-build
	mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) main.go

help:
	@echo "Usage: make [target]"
	@echo "Targets:"
	@echo "  clean          - Remove build artifacts"
	@echo "  test           - Run tests"
	@echo "  vet            - Run go vet"
	@echo "  tidy           - Run go mod tidy"
	@echo "  build          - Build the binary for the current OS"
	@echo "  build-windows  - Build the binary for Windows"
	@echo "  build-linux    - Build the binary for Linux"
	@echo "  build-macos    - Build the binary for macOS"