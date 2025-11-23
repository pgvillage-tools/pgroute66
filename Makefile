ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

build:
	go mod tidy
	go build ./cmd/pgroute66

debug:
	go build -gcflags "all=-N -l" ./cmd/pgroute66
	~/go/bin/dlv --headless --listen=:2345 --api-version=2 --accept-multiclient exec ./pgroute66 -- -c ./config/pgroute66_local.yaml

debug_traffic:
	curl 'http://localhost:8080/v1/primaries'
	curl 'http://localhost:8080/v1/primary'
	curl 'http://localhost:8080/v1/standbys'
	curl 'http://localhost:8080/v1/primaries?group=cluster'
	curl 'http://localhost:8080/v1/primary?group=cluster'
	curl 'http://localhost:8080/v1/standbys?group=cluster'

run:
	./pgroute66

fmt:
	gofmt -w .
	gofumpt -l -w .
	goimports -w .
	gci write .

lint:
	golangci-lint run -v

.PHONY: test
test: ## Run tests.
	go test $$(go list ./... | grep -v /e2e) -coverprofile cover.out

.PHONY: install-go-test-coverage
install-go-test-coverage:
	go install github.com/vladopajic/go-test-coverage/v2@latest

.PHONY: check-coverage
check-coverage: install-go-test-coverage
	go test $$(go list ./... | grep -v /e2e) -coverprofile=./cover.out -covermode=atomic -coverpkg=./...
	${GOBIN}/go-test-coverage --config=./.testcoverage.yaml
