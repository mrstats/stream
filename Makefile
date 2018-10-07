###
SHELL := /bin/bash

PLATFORM	= $(shell go env GOOS)
ARCH		= $(shell go env GOARCH)
DEF_ARCH	= ${ARCH}
GOPATH		= $(shell go env GOPATH)
GOBIN		= $(GOROOT)/bin/go

PWD			= $(shell pwd)
BUILD_DIR	= ${PWD}/build
CMD_DIR		= ./cmd/${BINARY}

###
default: clean dep queuer

###
build: ${PLATFORM}

linux:
	GOARCH=${DEF_ARCH} GOOS=linux   ${GOBIN} build -o ${BUILD_DIR}/linux/${BINARY}       ${CMD_DIR}

darwin:
	GOARCH=${DEF_ARCH} GOOS=darwin  ${GOBIN} build -o ${BUILD_DIR}/darwin/${BINARY}      ${CMD_DIR}

windows:
	GOARCH=${DEF_ARCH} GOOS=windows ${GOBIN} build -o ${BUILD_DIR}/windows/${BINARY}.exe ${CMD_DIR}

###
clean:
	-rm -f ${BUILD_DIR}/linux/${BINARY}*
	-rm -f ${BUILD_DIR}/darwin/${BINARY}*
	-rm -f ${BUILD_DIR}/windows/${BINARY}*

###
dep:
	go get

###
queuer: BINARY = queuer
queuer: build

###
.PHONY: default build clean linux darwin windows dep queuer
