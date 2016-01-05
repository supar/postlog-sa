.PHONY: test

NAME = postlog-sa
VERSION = $(shell cat VERSION)

BUILD_DIR = $(notdir $(shell pwd))
BUILD_DATE = $(shell date +%Y%m%d%H%M%S)

LDFLAGS = -ldflags "-X main.NAME=${NAME} -X main.VERSION=${VERSION} -X main.BUILDDATE=${BUILD_DATE}"

build:
	go build -o ./$(NAME) -v $(LDFLAGS)

test:
	@go test -v ./...

dependency:
	@go get -fix -t $(BUILD_PKGS)
