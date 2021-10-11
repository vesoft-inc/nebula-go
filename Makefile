.PHONY: build test fmt ci ssl-test run-examples

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

ssl-test:
	cd ./nebula-docker-compose && enable_ssl=true docker-compose up -d && \
	sleep 5 && \
	cd .. && \
	go test -v -run TestSslConnection; \
	cd ./nebula-docker-compose && docker-compose down -v 

ssl-test:
	cd ./nebula-docker-compose && enable_ssl=true docker-compose up -d && \
	sleep 5 && \
	cd .. && \
	ssl_test=true go test -v -run TestSslConnection; \
	cd ./nebula-docker-compose && docker-compose down -v 

run-examples:
	go run basic_example/graph_client_basic_example.go
	go run gorountines_example/graph_client_goroutines_example.go
