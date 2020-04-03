
GOPATH:=$(shell go env GOPATH)


.PHONY: build
build:

	CGO_ENABLED="0" GOOS="linux" go build -o route-web *.go

.PHONY: debug
debug:

	go build -o dbg-route-web *.go

.PHONY: test
test:
	go test -v ./... -cover

.PHONY: docker
docker: build
	docker build . -t 10.9.19.140/test/route-web:latest
