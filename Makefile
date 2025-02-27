.PHONY: build plugins clean

BINARY=grpcli
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date +%FT%T%z)
PLUGINS_DIR=plugins

build:
	go build -o $(BINARY) main.go

plugins:
	@mkdir -p $(PLUGINS_DIR)
	go build -buildmode=plugin -o $(PLUGINS_DIR)/kubernetes.so ./plugins/kubernetes/kubernetes.go

clean:
	rm -f $(BINARY)
	rm -f $(PLUGINS_DIR)/*.so

all: clean build plugins