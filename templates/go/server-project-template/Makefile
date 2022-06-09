PWD=$(shell pwd)
GO_VERSION=1.16.15
IMAGE_NAME=stella-go/go:${GO_VERSION}

.PHONY : all

all: .build-docker vendor
	@docker run --rm -v $(PWD):/src -w /src ${IMAGE_NAME} ./build.sh

clean:
	@rm -rf target vendor

.build-docker:
	@docker build --build-arg GO_VERSION=${GO_VERSION} -t ${IMAGE_NAME} . && touch .build-docker

vendor:
	go mod vendor