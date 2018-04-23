PROJECT_ROOT		:= github.com/infobloxopen/atlas-contacts-app
BUILD_PATH  		:= bin
DOCKERFILE_PATH		:= $(CURDIR)/docker

USERNAME		:= $(USER)
GIT_COMMIT 		:= $(shell git describe --dirty=-unsupported --always || echo pre-commit)
IMAGE_VERSION		?= v1.0

SERVER_BINARY 		:= $(BUILD_PATH)/server
SERVER_PATH 		:= $(PROJECT_ROOT)/cmd/server
SERVER_IMAGE		:= infoblox/atlas-contacts-app:$(IMAGE_VERSION)
SERVER_DOCKERFILE 	:= $(DOCKERFILE_PATH)/Dockerfile.contacts-server

GATEWAY_BINARY 		:= $(BUILD_PATH)/gateway
GATEWAY_PATH		:= $(PROJECT_ROOT)/cmd/gateway
GATEWAY_IMAGE		:= infoblox/atlas-contacts-app-gateway:$(IMAGE_VERSION)
GATEWAY_DOCKERFILE 	:= $(DOCKERFILE_PATH)/Dockerfile.contacts-gateway

GO_PATH              	:= /go
SRCROOT_ON_HOST      	:= $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
SRCROOT_IN_CONTAINER	:= $(GO_PATH)/src/$(PROJECT_ROOT)
GO_CACHE             	:= -pkgdir $(SRCROOT_IN_CONTAINER)/$(BUILD_PATH)/go-cache

DOCKER_RUNNER        	:= docker run --rm
DOCKER_RUNNER        	+= -v $(SRCROOT_ON_HOST):$(SRCROOT_IN_CONTAINER)
DOCKER_BUILDER       	:= infoblox/buildtool:v8
DOCKER_GENERATOR     	:= infoblox/atlas-gentool:v1
GENERATOR            	:= $(DOCKER_RUNNER) $(DOCKER_GENERATOR)

BUILD_TYPE ?= "default"
ifeq ($(BUILD_TYPE), "default")
	BUILDER        := $(DOCKER_RUNNER) -w $(SRCROOT_IN_CONTAINER) $(DOCKER_BUILDER)
endif

GO_BUILD_FLAGS		?= $(GO_CACHE) -i -v
GO_TEST_FLAGS		?= -v -cover
GO_TEST_PACKAGES	:= $(shell $(BUILDER) go list ./... | grep -v "./vendor/")
SEARCH_GOFILES		:= $(BUILDER) find . -not -path '*/vendor/*' -type f -name "*.go"

.PHONY: default
default: test server gateway

.PHONY: all
all: vendor protobuf test server gateway

.PHONY: fmt
fmt:
	@$(SEARCH_GOFILES) -exec gofmt -s -w {} \;

.PHONY: test
test: fmt
	@$(BUILDER) go test $(GO_TEST_FLAGS) $(GO_TEST_PACKAGES)

.PHONY: server
server: server-build server-docker

.PHONY: server-build
server-build:
	@$(BUILDER) go build $(GO_BUILD_FLAGS) -o $(SERVER_BINARY) $(SERVER_PATH)

.PHONY: server-docker
server-docker: server-build
	@docker build -f $(SERVER_DOCKERFILE) -t $(SERVER_IMAGE) .

.PHONY: gateway
gateway: gateway-build gateway-docker
	@$(BUILDER) go build $(GO_BUILD_FLAGS) -o $(SERVER_BINARY) $(SERVER_PATH)

.PHONY: gateway-build
gateway-build:
	@$(BUILDER) go build $(GO_BUILD_FLAGS) -o $(GATEWAY_BINARY) $(GATEWAY_PATH)

.PHONY: gateway-docker
gateway-docker:
	@docker build -f $(GATEWAY_DOCKERFILE) -t $(GATEWAY_IMAGE) .

.PHONY: protobuf
protobuf:
	@$(GENERATOR) \
	--go_out=plugins=grpc:. \
	--grpc-gateway_out=logtostderr=true:. \
	--validate_out="lang=go:." \
	--gorm_out=. \
	--swagger_out=:. $(PROJECT_ROOT)/proto/contacts.proto

.PHONY: vendor
vendor:
	$(BUILDER) dep ensure -vendor-only

.PHONY: vendor-update
vendor-update:
	$(BUILDER) dep ensure

.PHONY: image
image:
	docker build -f docker/Dockerfile.contacts-server -t infoblox/contacts-server:v1.0 .
	docker build -f docker/Dockerfile.contacts-gateway -t infoblox/contacts-gateway:v1.0 .

.PHONY: image-clean
image-clean:
	docker rmi -f infoblox/contacts-server:v1.0 infoblox/contacts-gateway:v1.0

.PHONY: up
up:
	kubectl apply -f kube/kube.yaml

.PHONY: down
down:
	kubectl delete -f kube/kube.yaml
