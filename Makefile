.PHONY: build test fmt ci run-examples

default: build

build: fmt
	go mod tidy 
	go build

test: 
	go mod tidy 
	go test -v -race

fmt:
	go fmt

ci:
	cd ./nebula-docker-compose && docker-compose up -d && \
	sleep 5 && \
	cd .. && \
	go test -v -race; \
	cd ./nebula-docker-compose && docker-compose down -v 

run-examples:
	go run basic_example/graph_client_basic_example.go
	go run gorountines_example/graph_client_goroutines_example.go
