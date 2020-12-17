.PHONY: build test test-dev

default: build

build: clean fmt
	go mod tidy 
	go build

test: 
	go mod tidy 
	go test
