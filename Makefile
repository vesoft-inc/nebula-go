.PHONY: build unit test e2e fmt run-examples

default: build

gen-code:
	rm -rf internal/generated_code/v5.0.0/proto/common
	rm -rf internal/generated_code/v5.0.0/proto/vector
	rm -rf internal/generated_code/v5.0.0/proto/graph
	cd proto && \
	protoc --go_out=. --go-grpc_out=. ./nebula/*.proto && \
	mv ./github.com/vesoft-inc/nebula-go/v5/internal/generated_code/v5.0.0/proto/* ../internal/generated_code/v5.0.0/proto/ && \
	rm -rf ./github.com

build: fmt
	go mod tidy
	go build

test:
	go mod tidy
	go list ./... |grep -v example|grep -v e2e|grep -v generated_code| xargs go test  -v -race -timeout 30s  --covermode=atomic  --coverprofile coverage.out

e2e-up:
	cd e2e/docker-compose && docker-compose pull && docker-compose up -d
	cd e2e/docker-compose-ssl && docker-compose pull && docker-compose up -d
	cd e2e/import && go run main.go

e2e-down:
	cd e2e/docker-compose && docker-compose down
	cd e2e/docker-compose-ssl && docker-compose down

e2e:
	 go test -v -race  --covermode=atomic  --coverprofile ./e2e.out  ./e2e/... --cover -coverpkg=./...

fmt:
	go fmt $(shell go list ./... | grep -v /generated_code/)

run-examples:
	go run examples/basic_example.go
