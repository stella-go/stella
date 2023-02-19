PWD=$(shell pwd)
GO_VERSION=1.19.6
IMAGE_NAME=stella-go/go:${GO_VERSION}
GOOS="linux" 
GOARCH="amd64"

.PHONY: all

all: .build-docker vendor
	@docker run --rm -v $(PWD):/src -w /src ${IMAGE_NAME} sh -c "GOOS=${GOOS} GOARCH=${GOARCH} ./build.sh"

clean:
	@rm -rf target vendor

.build-docker:
	@docker build  --memory-swap -1 --build-arg GO_VERSION=${GO_VERSION} -t ${IMAGE_NAME} . && echo > .build-docker

vendor:
	go mod vendor