
GOPATH:=$(shell go env GOPATH)


.PHONY: build
build:

	go build -o push-web *.go

.PHONY: test
test:
	go test -v ./... -cover

.PHONY: docker
docker:
	docker build . -t 10.9.19.140/test/push-web:latest
