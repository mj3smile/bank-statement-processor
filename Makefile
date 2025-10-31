PROJECT_NAME := bank-statement-processor
BUILD_DIR := bin
CMD_DIR := cmd/api

.PHONY: build
build:
	@echo "Building $(PROJECT_NAME)..."
	cd $(CMD_DIR) && go build -o ../../$(BUILD_DIR)/$(PROJECT_NAME) .

.PHONY: deps
deps:
	go mod download
	go mod verify

.PHONY: test
test:
	@echo "Running tests..."
	go test -race ./...

.PHONY: run
run: build
	@echo "running $(PROJECT_NAME)..."
	./$(BUILD_DIR)/$(PROJECT_NAME)