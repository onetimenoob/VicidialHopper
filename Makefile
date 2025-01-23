# Project name
PROJECT_NAME=VicidialHopper
# Go parameters
GO_CMD=go
GO_BUILD=$(GO_CMD) build
GO_TEST=$(GO_CMD) test
GO_CLEAN=$(GO_CMD) clean
# Build output directory
BUILD_DIR=bin
# Main source file
MAIN_SRC=main.go
# Binary output (now a Linux executable)
BINARY=$(BUILD_DIR)/$(PROJECT_NAME)
# Environment variables for cross-compilation and static linking
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64
# Default target
all: build
# Build target
build:
	$(GO_BUILD) -o $(BINARY) -ldflags "-w -s -extldflags '-static'" $(MAIN_SRC)
# Clean target
clean:
	$(GO_CLEAN)
	if exist $(BINARY) del $(BINARY)
	if exist $(BINARY).exe del $(BINARY).exe
# Test target
test:
	$(GO_TEST) ./...
# PHONY targets
.PHONY: all build clean test