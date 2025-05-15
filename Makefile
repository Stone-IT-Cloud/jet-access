.PHONY: all build clean test lint vet fmt help run

# Go parameters
BINARY_NAME=jet-access
MAIN_PACKAGE=./cmd/$(BINARY_NAME)
GOBUILD=go build
GOTEST=go test
GOLINT=golangci-lint
GOVET=go vet
GOFMT=gofmt
GOGET=go get
GOMOD=go mod

# Build directory
BUILD_DIR=build
BIN_DIR=$(BUILD_DIR)/bin

all: test build

build:
	mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(BIN_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out

test:
	$(GOTEST) -v ./...

test-coverage:
	$(GOTEST) -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

lint:
	$(GOLINT) run ./...

vet:
	$(GOVET) ./...

fmt:
	$(GOFMT) -l -w .

tidy:
	$(GOMOD) tidy

download:
	$(GOMOD) download

update:
	$(GOMOD) download all

run:
	go run $(MAIN_PACKAGE)

help:
	@echo "Available targets:"
	@echo "  all           - Run test and build"
	@echo "  build         - Build the application"
	@echo "  clean         - Remove build artifacts"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  lint          - Run linter"
	@echo "  vet           - Run go vet"
	@echo "  fmt           - Run gofmt"
	@echo "  tidy          - Run go mod tidy"
	@echo "  download      - Download dependencies"
	@echo "  update        - Update dependencies"
	@echo "  run           - Run the application"
	@echo "  help          - Display this help message"

# Default target
.DEFAULT_GOAL := help
