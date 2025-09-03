.PHONY: all build build-mcp run login tidy fmt vet test clean help

BIN_DIR := bin

all: build

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/chatbang ./cmd/chatbang

build-mcp:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/mcp ./cmd/mcp

run:
	go run ./cmd/chatbang

login:
	go run ./cmd/chatbang login

tidy:
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test ./...

clean:
	rm -rf $(BIN_DIR)

help:
	@echo "Targets:"
	@echo "  build       - Build chatbang binary to $(BIN_DIR)/chatbang"
	@echo "  build-mcp   - Build mcp provider binary to $(BIN_DIR)/mcp"
	@echo "  run         - Run chatbang"
	@echo "  login       - Run chatbang login"
	@echo "  tidy        - go mod tidy"
	@echo "  fmt         - go fmt ./..."
	@echo "  vet         - go vet ./..."
	@echo "  test        - go test ./..."
	@echo "  clean       - Remove $(BIN_DIR)"

