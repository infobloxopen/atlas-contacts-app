PROJECT_ROOT		:= github.com/infobloxopen/atlas-contacts-app
BUILD_PATH  		:= bin
DOCKERFILE_PATH		:= $(CURDIR)/docker

# configuration for image names
USERNAME        := $(USER)
GIT_COMMIT      := $(shell git describe --dirty=-unsupported --always || echo pre-commit)
IMAGE_VERSION   ?= $(USERNAME)-dev-$(GIT_COMMIT)
IMAGE_REGISTRY  ?= infoblox

# configuration for server binary and image
SERVER_BINARY 		:= $(BUILD_PATH)/server
SERVER_PATH 			:= $(PROJECT_ROOT)/cmd/server
SERVER_IMAGE			:= $(IMAGE_REGISTRY)/contacts-server
SERVER_DOCKERFILE := $(DOCKERFILE_PATH)/Dockerfile.server

# configuration for gateway binary and image
GATEWAY_BINARY 	   := $(BUILD_PATH)/gateway
GATEWAY_PATH		   := $(PROJECT_ROOT)/cmd/gateway
GATEWAY_IMAGE		   := $(IMAGE_REGISTRY)/contacts-gateway
GATEWAY_DOCKERFILE := $(DOCKERFILE_PATH)/Dockerfile.gateway

# configuration for the protobuf gentool
SRCROOT_ON_HOST		:= $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
SRCROOT_IN_CONTAINER	:= /go/src/$(PROJECT_ROOT)
DOCKER_RUNNER        	:= docker run --rm
DOCKER_RUNNER        	+= -v $(SRCROOT_ON_HOST):$(SRCROOT_IN_CONTAINER)
DOCKER_GENERATOR	:= infoblox/atlas-gentool:v3
GENERATOR		:= $(DOCKER_RUNNER) $(DOCKER_GENERATOR)

# configuration for the database
DATABASE_ADDRESS		?= localhost:5432

# configuration for building on host machine
GO_CACHE		:= -pkgdir $(BUILD_PATH)/go-cache
GO_BUILD_FLAGS		?= $(GO_CACHE) -i -v
GO_TEST_FLAGS		?= -v -cover
GO_PACKAGES		:= $(shell go list ./... | grep -v vendor)

.PHONY: all
all: vendor protobuf server-docker gateway-docker

.PHONY: fmt
fmt:
	@go fmt $(GO_PACKAGES)

.PHONY: test
test: fmt
	@go test $(GO_TEST_FLAGS) $(GO_PACKAGES)

.PHONY: server-docker
server-docker:
	@docker build -f $(SERVER_DOCKERFILE) -t $(SERVER_IMAGE):$(IMAGE_VERSION) .

.PHONY: gateway-docker
gateway-docker:
	@docker build -f $(GATEWAY_DOCKERFILE) -t $(GATEWAY_IMAGE):$(IMAGE_VERSION) .

.PHONY: push
push:
	@docker push $(SERVER_IMAGE)
	@docker push $(GATEWAY_IMAGE)

.PHONY: protobuf
protobuf:
	@$(GENERATOR) \
	--go_out=plugins=grpc:. \
	--grpc-gateway_out=logtostderr=true:. \
	--gorm_out=. \
	--validate_out="lang=go:." \
	--swagger_out=:. $(PROJECT_ROOT)/pkg/pb/contacts.proto

.PHONY: vendor
vendor:
	@dep ensure -vendor-only

.PHONY: vendor-update
vendor-update:
	@dep ensure

.PHONY: clean
clean:
	@docker rmi -f $(shell docker images -q $(SERVER_IMAGE)) || true
	@docker rmi -f $(shell docker images -q $(GATEWAY_IMAGE)) || true
	@docker rmi -f $(shell docker images --filter "label=intermediate=true" -q) || true

.PHONY: migrate-up
migrate-up:
	@migrate -database 'postgres://$(DATABASE_ADDRESS)/contacts?sslmode=disable' -path ./db/migrations up

.PHONY: migrate-down
migrate-down:
	@migrate -database 'postgres://$(DATABASE_ADDRESS)/contacts?sslmode=disable' -path ./db/migrations down

.PHONY: up
up:
	kubectl apply -f deploy/kube.yaml

.PHONY: down
down:
	kubectl delete -f deploy/kube.yaml

.PHONY: nginx-up
nginx-up:
	kubectl apply -f deploy/nginx.yaml

.PHONY: nginx-down
nginx-down:
	kubectl delete -f deploy/nginx.yaml
