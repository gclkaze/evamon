APP_NAME  := evamon
MAIN_FILE := ./main.go
BUILD_DIR := ./build
OUT       := $(BUILD_DIR)/$(APP_NAME)

.PHONY: all build test clean fmt vet tidy

all: build

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(OUT) $(MAIN_FILE)

test:
	go test ./... -count=1

clean:
	rm -rf $(BUILD_DIR)

fmt:
	go fmt ./...

vet:
	go vet ./...

tidy:
	go mod tidy
