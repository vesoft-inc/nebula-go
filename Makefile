.PHONY: build unit test fmt up up-ssl down ssl-test

default: build

build: fmt
	go mod tidy
	go build
unit:
	go mod tidy
	go test -v -race --covermode=atomic --coverprofile coverage.out
test:
	go mod tidy
	go test -v -race --tags=integration --covermode=atomic --coverprofile coverage.out

fmt:
	go fmt

lint:
	@test -z `gofmt -l *.go` || (echo "Please run 'make fmt' to format Go code" && exit 1)

up:
	cd ./nebula-docker-compose && docker compose up -d

up-ssl:
	cd ./nebula-docker-compose && enable_ssl=true docker compose -f docker-compose-ssl.yaml up -d

down:
	cd ./nebula-docker-compose && docker compose down -v

ssl-test:
	ssl_test=true go test -v --tags=integration -run TestSslConnection;
	ssl_test=true go test -v --tags=integration -run TestSslSessionPool;

ssl-test-self-signed:
	self_signed=true go test -v --tags=integration -run TestSslConnection;
