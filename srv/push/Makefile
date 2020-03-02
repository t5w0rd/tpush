
GOPATH:=$(shell go env GOPATH)


.PHONY: proto
proto:
    
	protoc --proto_path=. --micro_out=. --go_out=. proto/push/push.proto
    

.PHONY: build
build: proto

	CGO_ENABLED="0" GOOS="linux" go build -o push-srv *.go

.PHONY: test
test:
	go test -v ./... -cover

.PHONY: docker
docker:
	docker build . -t 10.9.19.140/test/push-srv:latest
