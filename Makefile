###
SHELL := /bin/bash

#PLATFORM	= linux
PLATFORM	= $(shell go env GOOS)
ARCH		= $(shell go env GOARCH)
DEF_ARCH	= ${ARCH}
GOPATH		= $(shell go env GOPATH)
GOBIN		= $(GOROOT)/bin/go
CGO			= 0

PWD			= $(shell pwd)
BUILD_DIR	= ${PWD}/deployments/build
CMD_DIR		= ./cmd/${BINARY}

###
default: clean dep queuer wsocket storage docker

###
build:
	@echo "building ${BINARY} for ${PLATFORM}..."
	@GOOS=${PLATFORM} CGO_ENABLED=${CGO} GOARCH=${DEF_ARCH} ${GOBIN} build \
		-a \
		-installsuffix cgo \
		-o ${BUILD_DIR}/${PLATFORM}/${BINARY} \
		${CMD_DIR}

###
clean:
	@echo "removing all binaries..."
	@-rm -rf ${BUILD_DIR}/*

###
dep:
	@echo "geting dependncies..."
	@go get

###
queuer:
	@echo "make queuer:"
	@$(MAKE) BINARY="queuer"	build

###
wsocket:
	@echo "make wsocket:"
	@$(MAKE) BINARY="wsocket"	build

###
storage:
	@echo "make storage:"
	@echo "not implimented"
#	@$(MAKE) BINARY="storage"	build


###
docker:
	@echo "docker compose:"
	docker-compose -f deployments/docker-compose.yml up
###
.PHONY: default build clean linux darwin windows dep queuer wsocket storage docker
