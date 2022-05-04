.PHONY: build test fmt up up-ssl down ssl-test run-examples

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

up-ssl:
	cd ./nebula-docker-compose && enable_ssl=true docker-compose -f docker-compose-ssl.yaml up -d

down:
	cd ./nebula-docker-compose && docker-compose down -v

ssl-test:
	ssl_test=true go test -v -run TestSslConnection;

ssl-test-self-signed:
	self_signed=true go test -v -run TestSslConnection;

run-examples:
	go run examples/basic_example/graph_client_basic_example.go && \
	go run examples/basic_example/parameter_example.go && \
	go run examples/gorountines_example/graph_client_goroutines_example.go && \
	go run examples/json_example/parse_json_example.go

.PHONY: clean coverage lint gofumpt goimports gomod_tidy govet staticcheck golangci-lint golangci-lint-fix

clean:
	rm coverage.*

coverage:
	$(info should install from github.com/t-yuki/gocover-cobertura)
	gocover-cobertura < coverage.out > coverage.xml

lint: fmt gofumpt goimports gomod_tidy govet golangci-lint-fix staticcheck
	$(info running all linters)

gofumpt:
	$(info install via go get -u/ go install + mvdan.cc/gofumpt@latest)
	gofumpt -l -w -extra *.go

goimports:
	$(info install via go get -u/ go install + golang.org/x/tools/cmd/goimports@latest)
	goimports -e -l -w -local github.com *.go

gomod_tidy:
	$(info add missing and remove unused modules)
	go mod tidy

govet:
	$(info Vet examines Go source code and reports suspicious constructs. See https://pkg.go.dev/cmd/vet)
	go vet -all

staticcheck:
	$(info install via go get -u / go install + honnef.co/go/tools/cmd/staticcheck@latest)
	staticcheck

GOLANGCI_LINT_FLAGS = -v --max-same-issues 0 --max-issues-per-linter 0 --deadline=300s

golangci-lint:
	$(info running local golangci-lint, to install check https://golangci-lint.run/usage/install/)
	golangci-lint run $(GOLANGCI_LINT_FLAGS)

golangci-lint-fix:
	$(info running golangci-lint with --fix option, to install check https://golangci-lint.run/usage/install/)
	golangci-lint run $(GOLANGCI_LINT_FLAGS) --fix

.PHONY: docker-golangci-lint

DOCKER_GOLANGCI_LINT_CMD = docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:latest golangci-lint

docker-golangci-lint:
	$(info running golangci-lint via docker)
	$(DOCKER_GOLANGCI_LINT_CMD) run $(GOLANGCI_LINT_FLAGS)

