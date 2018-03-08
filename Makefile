# Absolute github repository name.
REPO := github.com/infobloxopen/atlas-contacts-app

# Build directory absolute path.
BINDIR = $(CURDIR)

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
	@docker run --rm -v ~:/root -v $(CURDIR):/go/src/$(REPO) -w /go/src/$(REPO) $(BUILDTOOL_IMAGE) \
	/bin/sh -c 'go get $(REPO)/cmd/gateway && go build $(GO_BUILD_FLAGS) -o "bin/gateway" "$(REPO)/cmd/gateway"'

	@docker run --rm -v $(CURDIR):/go/src/$(REPO) -w /go/src/$(REPO) $(BUILDTOOL_IMAGE) \
	/bin/sh -c 'go get $(REPO)/cmd/contacts && go build $(GO_BUILD_FLAGS) -o "bin/contacts" "$(REPO)/cmd/contacts"'

clean:
	@rm -rf "$(BINDIR)"

fmt:
	@echo "Running 'go fmt ...'"
	@go fmt -x "$(REPO)/..."

image: build
	cd docker && docker build -f Dockerfile.contacts -t contacts:latest .
	cd docker && docker build -f Dockerfile.gateway -t gateway:latest .

up: image
	@docker network create exampleexternal

	@docker run --name tagging -d -p "9091:91" \
		--network exampleexternal \
		--network-alias tagging.exampleexternal \
		tagging:latest \
			--listen 0.0.0.0:91

	@docker run --name dnsconfig -d -p "9092:92" \
		--network exampleexternal \
		--network-alias dnsconfig.exampleexternal \
		dnsconfig:latest \
			--listen 0.0.0.0:92

	@docker run --name gateway -d -p "8080:80" \
		--network exampleexternal \
		--network-alias gateway.exampleexternal \
		gateway:latest \
			--listen 0.0.0.0:80 \
			--tagging tagging.exampleexternal:91 \
			--dnsconfig dnsconfig.exampleexternal:92

	@docker network inspect exampleexternal

down:
	@docker stop tagging
	@docker rm -f tagging

	@docker stop dnsconfig
	@docker rm -f dnsconfig

	@docker stop gateway
	@docker rm -f gateway

	@docker network rm exampleexternal

.PHONY:
	default clean build fmt gen
