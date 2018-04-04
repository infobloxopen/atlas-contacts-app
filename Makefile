# Absolute github repository name.
REPO := github.com/infobloxopen/atlas-contacts-app
DOCKER := johnbelamaric/atlas-contacts-app

# Build directory absolute path.
BINDIR = $(CURDIR)/bin

# Utility docker image to build Go binaries
# https://github.com/infobloxopen/buildtool
BUILDTOOL_IMAGE := infoblox/buildtool:latest

# Utility docker image to generate Go files from .proto definition.
# https://github.com/Infoblox-CTO/ngp.api.toolkit/gentool
GENTOOL_IMAGE := infoblox/gentool:latest

default: build

gen:
	@docker run -v $(CURDIR):/go/src/$(REPO) $(GENTOOL_IMAGE) \
	--go_out=plugins=grpc:. \
	--grpc-gateway_out=logtostderr=true:. \
	--validate_out="lang=go:." \
	--gorm_out="." \
	--swagger_out=:. $(REPO)/proto/contacts.proto

build:
	mkdir -p bin
	@docker run --rm -v ~:/root -v $(CURDIR):/go/src/$(REPO) -w /go/src/$(REPO) $(BUILDTOOL_IMAGE) \
	go build $(GO_BUILD_FLAGS) -o "bin/gateway" "$(REPO)/cmd/gateway"

	@docker run --rm -v $(CURDIR):/go/src/$(REPO) -w /go/src/$(REPO) $(BUILDTOOL_IMAGE) \
	go build $(GO_BUILD_FLAGS) -o "bin/contacts" "$(REPO)/cmd/contacts"

clean:
	@rm -rf "$(BINDIR)"

fmt:
	@echo "Running 'go fmt ...'"
	@go fmt -x "$(REPO)/..."

image: build
	docker build -f docker/Dockerfile.contacts -t $(DOCKER)-contacts:latest .
	docker build -f docker/Dockerfile.gateway -t $(DOCKER)-gateway:latest .

push: image
	@docker push $(DOCKER)-contacts:latest
	@docker push $(DOCKER)-gateway:latest

.PHONY:
	default clean build fmt gen
