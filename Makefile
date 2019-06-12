NAME = docker-machine-driver-smtxos
REPO = github.com/smartxworks/$(NAME)
VERSION = $(shell git describe --tags --dirty | awk '{print substr($$1,2)}')

.PHONY: all
all: build tar

.PHONY: fmt
fmt:
	go fmt $$(go list ./... | grep -v /vendor/)

.PHONY: vendor
vendor:
	go mod tidy
	go mod vendor

.PHONY: build
build:
	docker run --privileged --rm \
		-v $(shell pwd):/go/src/$(REPO) -w /go/src/$(REPO) \
		golang:1.12 make _build

.PHONY: _build
_build:
	GOOS=linux GOARCH=amd64 go build .

.PHONY: tar
tar:
	tar zcvf $(NAME)-$(VERSION).linux-amd64.tar.gz $(NAME)
