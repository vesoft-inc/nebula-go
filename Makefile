.PHONY: build test fmt up down ci ssl-test run-examples

default: build

build: fmt
	go mod tidy
	go build

test:
	go mod tidy
	go test -v -race

fmt:
	go fmt
up:
	cd ./nebula-docker-compose && docker-compose up -d

down:
	cd ./nebula-docker-compose && docker-compose down -v

ci:
	go test -v -race;

ssl-test:
	go test -v -run TestSslConnection;

ssl-test-self-signed:
	cd ./nebula-docker-compose && enable_ssl=true docker-compose up -d && \
	sleep 10 && \
	cd .. && \
	ssl_test=true go test -v -run TestSslConnection;

run-examples:
	go run basic_example/graph_client_basic_example.go && \
	go run gorountines_example/graph_client_goroutines_example.go && \
	go run json_example/parse_json_example.go
