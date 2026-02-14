# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOLINT=bin/golangci-lint run
BINARY_NAME=eusurveymgr
BUILD_DIR=bin

DATE=$(shell date +%Y%m%d_%H%M%S)
VERSION=0.1.0
COMMIT=$(shell git rev-parse HEAD)
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildDate=$(DATE)"

all: darwin-arm64 linux-amd64 windows-amd64

darwin-arm64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME).darwin-arm64.bin

darwin-amd64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME).darwin-amd64.bin

linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME).linux-amd64.bin

windows-amd64:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME).windows-amd64.exe

test:
	$(GOTEST) -v ./...

lint:
	$(GOLINT) ./...

cover:
	$(GOCMD) test -cover ./...

-include Makefile.local

clean:
	$(GOCLEAN)
	rm -f $(BUILD_DIR)/*.bin $(BUILD_DIR)/*.exe