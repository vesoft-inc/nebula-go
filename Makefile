.PHONY: build test up down fmt ssl-test ssl-test-ca-signed ssl-test-self-signed run-examples

default: build

build: fmt
	go mod tidy
	go build

test: fmt
	go mod tidy
	go test -v -race

up:
	cd ./nebula-docker-compose && docker-compose up -d && \
	sleep 15

down:
	cd ./nebula-docker-compose && docker-compose down -v

fmt:
	go fmt

ssl-test-ca-signed:
	cd ./nebula-docker-compose && enable_ssl=true \
	password_path="" \
	docker-compose up -d && \
	sleep 15 && \
	cd .. && \
	ca_signed=true go test -v -run TestSslConnectionCaSigned

ssl-test-self-signed:
	cd ./nebula-docker-compose && enable_ssl=true docker-compose up -d && \
	sleep 15 && \
	cd .. && \
	self_signed=true go test -v -run TestSslConnectionSelfSigned

run-examples:
	go run basic_example/graph_client_basic_example.go && \
	go run gorountines_example/graph_client_goroutines_example.go && \
	go run json_example/parse_json_example.go
