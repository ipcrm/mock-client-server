default: ci

ci: lint fmt-check imports-check

# Tooling versions
GOXVERSION?=v1.0.1
CGO_ENABLED?=0
export CGO_ENABLED

.PHONY: prepare
prepare: install-tools 

.PHONY: lint
lint: ## Runs go linter
	golangci-lint run

.PHONY: fmt
fmt: ## Runs and applies go formatting changes
	@gofmt -w -l $(shell go list -f {{.Dir}} ./...)
	@goimports -w -l $(shell go list -f {{.Dir}} ./...)

.PHONY: fmt-check
fmt-check: ## Lists formatting issues
	@test -z $(shell gofmt -l $(shell go list -f {{.Dir}} ./...))

.PHONY: imports-check
imports-check: ## Lists imports issues
	@test -z $(shell goimports -l $(shell go list -f {{.Dir}} ./...))

.PHONY: build
build: build-server-cross-platform build-client-cross-platform

.PHONY: build-server-cross-platform
build-server-cross-platform: ## Compiles the Server binary for all supported platforms
	gox -output="bin/server-{{.OS}}-{{.Arch}}" \
            -os="linux" \
            -arch="amd64 386" \
            -osarch="darwin/amd64 darwin/arm64 linux/arm linux/arm64" \
            github.com/ipcrm/mock-client-server/server

.PHONY: build-client-cross-platform
build-client-cross-platform: ## Compiles the client binary for all supported platforms
	gox -output="bin/client-{{.OS}}-{{.Arch}}" \
            -os="linux" \
            -arch="amd64 386" \
            -osarch="darwin/amd64 darwin/arm64 linux/arm linux/arm64" \
            github.com/ipcrm/mock-client-server/client


.PHONY: install-tools
install-tools: ## Install go indirect dependencies
ifeq (, $(shell which golangci-lint))
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v$(GOLANGCILINTVERSION)
endif
ifeq (, $(shell which goimports))
	GOFLAGS=-mod=readonly go install golang.org/x/tools/cmd/goimports@$(GOIMPORTSVERSION)
endif
ifeq (, $(shell which gox))
	GOFLAGS=-mod=readonly go install github.com/mitchellh/gox@$(GOXVERSION)
endif
ifeq (, $(shell which gotestsum))
	GOFLAGS=-mod=readonly go install gotest.tools/gotestsum@$(GOTESTSUMVERSION)
endif
