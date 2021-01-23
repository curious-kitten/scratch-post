SHELL:=/bin/bash
TOP_DIR:=$(notdir $(CURDIR))
BUILD_DIR:=build
BIN_DIR:=$(BUILD_DIR)/_bin

DOCKER_REPO?="matache91mh"
APP:=scratch-post
IMAGE?=$(DOCKER_REPO)/$(APP)

DB_CONF_FILE?=dbconfig.json
API_CONF_FILE?=apiconfig.json


ifeq ($(VERSION),)
	VERSION:=$(shell git describe --tags --dirty --always)
endif


all: install-go-tools lint run-tests build
	
build-app: lint
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -v -o $(BIN_DIR)/$(APP) ./cmd/$(APP)

test:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

install-go-tools:
	@./scripts/install_tools.sh
	go install github.com/golang/mock/mockgen

lint: fmt
	golangci-lint run ./...

generate:
	go generate -v ./...

run: build-app
	$(BIN_DIR)/$(APP) --apiconfig $(API_CONF_FILE) --dbconfig $(DB_CONF_FILE)


app-image: build-app
	cp $(BIN_DIR)/$(APP) $(BUILD_DIR)/$(APP)/ && \
	docker build -t $(IMAGE):$(VERSION) $(BUILD_DIR)/$(APP)/ && \
	rm  $(BUILD_DIR)/$(APP)/$(APP) && 

push-images: app-image
	docker push $(IMAGE):$(VERSION)

fmt:
	go mod tidy
	goimports -w .
	gofmt -s -w .

