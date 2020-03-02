
GOPATH:=$(shell go env GOPATH)


.PHONY: proto
proto:
    
	protoc --proto_path=. --micro_out=. --go_out=. proto/gateway/gateway.proto
    

.PHONY: build
build: proto

	CGO_ENABLED="0" GOOS="linux" go build -o gateway-srv *.go

.PHONY: test
test:
	go test -v ./... -cover

.PHONY: docker
docker:
	docker build . -t 10.9.19.140/test/gateway-srv:latest
