SHELL:=/bin/bash
TOP_DIR:=$(notdir $(CURDIR))
APP:=scratch-post
BUILD_DIR:=build
BIN_DIR:=$(BUILD_DIR)/$(APP)/_bin

DOCKER_REPO?="matache91mh"
IMAGE?=$(DOCKER_REPO)/$(APP)

ADMIN_DB_CONF_FILE?=admindb.json
TEST_DB_CONF_FILE?=testdb.json
API_CONF_FILE?=apiconfig.json

VERSION ?= $(shell git describe --tags --dirty --always)
BUILD_DATE ?= $(shell date +%FT%T%z)
COMMIT_HASH ?= $(shell git rev-parse --short HEAD 2>/dev/null)

LDFLAGS += -X 'github.com/curious-kitten/scratch-post/internal/info.version=${VERSION}'
LDFLAGS += -X 'github.com/curious-kitten/scratch-post/internal/info.commitHash=${COMMIT_HASH}'
LDFLAGS += -X 'github.com/curious-kitten/scratch-post/internal/info.buildDate=${BUILD_DATE}'


all: install-go-tools generate fmt lint test build-app
	
build-app:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -v -o $(BIN_DIR)/$(APP) ./cmd/$(APP)

test:
	go test -v  -json -coverprofile=coverage.out ./... > unit-test.json
	go tool cover -func=coverage.out

install-go-tools:
	GO111MODULE=on CGO_ENABLED=0 go get github.com/golangci/golangci-lint/cmd/golangci-lint
	go install github.com/golang/mock/mockgen
	go get golang.org/x/tools/cmd/goimports

lint:
	go vet ./...
	golangci-lint run ./...

generate: generate-proto
	go generate -v ./...

run-jwt: build-app
	$(BIN_DIR)/$(APP) --apiconfig $(API_CONF_FILE) --admindb $(ADMIN_DB_CONF_FILE) --testdb $(TEST_DB_CONF_FILE) --isJWT

run: 
	GO111MODULE=on GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go run -ldflags="$(LDFLAGS)" ./cmd/$(APP) start --apiconfig $(API_CONF_FILE) --admindb $(ADMIN_DB_CONF_FILE) --testdb $(TEST_DB_CONF_FILE)


app-image: build-app
	docker build -t $(IMAGE):$(VERSION) $(BUILD_DIR)/$(APP)

push-images: app-image
	docker push $(IMAGE):$(VERSION)

fmt:
	go mod tidy
	goimports -w .
	gofmt -s -w .

generate-proto:
	protoc -I=api/v1/metadata --go_out=pkg/api/v1/metadata/  --go_opt=paths=source_relative --doc_out=./docs/proto --doc_opt=markdown,metadata.md api/v1/metadata/*.proto
	protoc --proto_path=api/v1/scenario --proto_path=api/v1/  --go_out=pkg/api/v1/scenario/  --go_opt=paths=source_relative --doc_out=./docs/proto --doc_opt=markdown,scenario.md api/v1/scenario/*.proto
	protoc --proto_path=api/v1/testplan --proto_path=api/v1/  --go_out=pkg/api/v1/testplan/  --go_opt=paths=source_relative --doc_out=./docs/proto --doc_opt=markdown,testplan.md api/v1/testplan/*.proto
	protoc --proto_path=api/v1/project --proto_path=api/v1/  --go_out=pkg/api/v1/project/  --go_opt=paths=source_relative --doc_out=./docs/proto --doc_opt=markdown,project.md api/v1/project/*.proto
	protoc --proto_path=api/v1/execution --proto_path=api/v1/  --go_out=pkg/api/v1/execution/  --go_opt=paths=source_relative --doc_out=./docs/proto --doc_opt=markdown,execution.md api/v1/execution/*.proto
