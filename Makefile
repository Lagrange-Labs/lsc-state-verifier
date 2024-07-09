VERSION := $(shell git describe --tags --always)
GITREV := $(shell git rev-parse --short HEAD)
GITBRANCH := $(shell git rev-parse --abbrev-ref HEAD)
DATE := $(shell LANG=US date +"%a, %d %b %Y %X %z")

GOBASE := $(shell pwd)
GOBIN := $(GOBASE)/dist
GOARCH := $(ARCH)
GOENVVARS := GOBIN=$(GOBIN) CGO_ENABLED=1 GOOS=$(OS) GOARCH=$(GOARCH)
GOBINARY := lsc-state-verifier
GOCMD := $(GOBASE)/cmd/

LDFLAGS += -X 'github.com/Lagrange-Labs/lagrange-state-verifier.Version=$(VERSION)'
LDFLAGS += -X 'github.com/Lagrange-Labs/lagrange-state-verifier.GitRev=$(GITREV)'
LDFLAGS += -X 'github.com/Lagrange-Labs/lagrange-state-verifier.GitBranch=$(GITBRANCH)'
LDFLAGS += -X 'github.com/Lagrange-Labs/lagrange-state-verifier.BuildDate=$(DATE)'

STOP := docker compose down --remove-orphans

# Building the docker image and the binary
build: ## Builds the binary locally into ./dist
	$(GOENVVARS) go build -ldflags "all=$(LDFLAGS)" -o $(GOBIN)/$(GOBINARY) $(GOCMD)
.PHONY: build

docker-build: ## Builds a docker image with the binary
	docker build -t lsc-state-verifier -f ./Dockerfile .
.PHONY: docker-build

# Linting, Teseting, Benchmarking
golangci_lint_cmd=github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2

install-linter:
	@echo "--> Installing linter"
	@go install $(golangci_lint_cmd)

lint:
	@echo "--> Running linter"
	@ $$(go env GOPATH)/bin/golangci-lint run --timeout=10m
.PHONY:	lint install-linter

test: 
	go test ./... --timeout=10m -v --race
.PHONY: test
run: build
	./dist/$(GOBINARY) db

localnet-start: stop
	echo "Starting localnet"
	docker compose up

stop:
	$(STOP)
.PHONY: localnet-start stop
