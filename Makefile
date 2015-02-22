all: install

deps:
	go get -v -d ./...

install: deps
	go install ./...

build: deps
	go build -o test1 main.go

test1: main.go
	docker run --rm -v "$$PWD":/usr/src/myapp -w /usr/src/myapp golang:1.4 make build

docker: test1
	docker build -t smira/test1 .
	docker push smira/test1

.PHONY: install build deps docker
