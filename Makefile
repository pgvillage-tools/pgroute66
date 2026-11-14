ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

VERSION ?= v1.2.3

build:
	go mod tidy
	go build -ldflags="-X 'github.com/pgvillage-tools/pgroute66/internal/version.appVersion=${VERSION}'" -o pgroute66 ./cmd/pgroute66

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

# Same version as the go-test-coverage action in .github/workflows/ci-run-on-pr.yaml: Dependabot bumps that one,
# update this one with it. Since v2.20.0 it needs Go 1.27 to install (GOTOOLCHAIN=auto downloads it).
GO_TEST_COVERAGE_VERSION ?= v2.23.0

.PHONY: install-go-test-coverage
install-go-test-coverage:
	go install github.com/vladopajic/go-test-coverage/v2@$(GO_TEST_COVERAGE_VERSION)

.PHONY: check-coverage
check-coverage: install-go-test-coverage
	go test $$(go list ./... | grep -v /e2e) -coverprofile=./cover.out -covermode=atomic -coverpkg=./...
	${GOBIN}/go-test-coverage --config=./.testcoverage.yaml
