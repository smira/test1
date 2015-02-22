all: install

deps:
	go get -v -d ./...

install: deps
	go install ./...


.PHONY: install deps