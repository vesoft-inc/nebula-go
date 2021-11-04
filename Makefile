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
	sleep 15 && \
	cd .. && \
	go test -v -race; \
	cd ./nebula-docker-compose && docker-compose down -v

ssl-test-ca-signed:
	cd ./nebula-docker-compose && enable_ssl=true \
	password_path="" \
	docker-compose up -d && \
	sleep 15 && \
	cd .. && \
	ca_signed=true go test -v -run TestSslConnectionCaSigned; \
	cd ./nebula-docker-compose && docker-compose down -v

ssl-test-self-signed:
	cd ./nebula-docker-compose && enable_ssl=true docker-compose up -d && \
	sleep 15 && \
	cd .. && \
	self_signed=true go test -v -run TestSslConnectionSelfSigned; \
	cd ./nebula-docker-compose && docker-compose down -v

run-examples:
	cd ./nebula-docker-compose && enable_ssl=true docker-compose up -d && \
	sleep 15 && \
	cd .. && \
	go run basic_example/graph_client_basic_example.go && \
	go run gorountines_example/graph_client_goroutines_example.go && \
	go run json_example/parse_json_example.go && \
	cd ./nebula-docker-compose && docker-compose down -v
