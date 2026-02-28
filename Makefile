.ONESHELL:
SHELL := /bin/bash

.SILENT:

UNAME_OS := $(shell uname -s)
UNAME_ARCH := $(shell uname -m)

ifeq ($(UNAME_OS), Darwin)
    GO_OS := darwin
else ifeq ($(UNAME_OS), Linux)
    GO_OS := linux
else
    $(error Unsupported operating system: $(UNAME_OS))
endif

ifeq ($(UNAME_ARCH), x86_64)
    GO_ARCH := amd64
else ifeq ($(UNAME_ARCH), arm64)
    GO_ARCH := arm64
else
    $(error Unsupported architecture: $(UNAME_ARCH))
endif

# Needed until native support will be implemented: https://github.com/golang/go/issues/37475
### [version] [branch] revision[-dirty] build_date_time
GITVER = $(shell \
  ver=$$(git tag -l --sort=-version:refname --merged HEAD 'v*' | head -n 1); \
  branch=$$(git rev-parse --abbrev-ref HEAD); \
  rev=$$(git log -1 --format='%h'); \
  git update-index -q --refresh --unmerged >/dev/null; \
  git diff-index --quiet HEAD || dirty="-dirty"; \
  test "$$branch" = "HEAD" || test "$$branch" = "master" && branch=; \
  echo "$${ver:+$$ver }$${branch:+$$branch }$$rev$$dirty $$(date -u +"%F_%T")" \
)

module = $(shell go list -m)

STATIC = false
EXTRA_ARGS =

ifndef CGO_ENABLED
	export CGO_ENABLED=0
endif

ifndef GOOS
	export GOOS=$(GO_OS)
endif

ifndef GOARCH
	export GOARCH=$(GO_ARCH)
endif

%:
	@:

DEFAULT_GOAL := help
.PHONY: help
help:
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*#"; printf "\n"} /^([a-zA-Z_-]+):.*?#/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: clean 
clean: # cleans directory with compiled binaries
	rm -rf bin/*

.PHONY: build
build: clean # compiles applications to statically/dynamically linked binaries
	@echo "Building the application..."
ifneq ($(STATIC),false)
	go build -ldflags "-w -extldflags '-static' -X '$(module)/pkg/def.ver=$(GITVER)'" -o bin/ $(EXTRA_ARGS) ./cmd/*

else
	go build -ldflags "-X '$(module)/pkg/def.ver=$(GITVER)'" -o bin/ $(EXTRA_ARGS) ./cmd/*
endif

.PHONY: run
run: # just executes main file using go, w/o separate compilation (for local testing only)
	go run ./cmd/server/main.go

.PHONY: static
static: # compiles app to statically linked binary
	$(MAKE) build STATIC=true

.PHONY: test
test: # runs all available tests
	CGO_ENABLED=1 go test -v -coverprofile=coverage.out -race -count=1 -timeout 30s $$(go list ./... | grep -v '/cmd')
	go tool cover -func=coverage.out | grep total

.PHONY: docker.build
docker.build: # builds docker container
	docker build -t ghcr.io/example/service:latest .

.PHONY: docker.run
docker.run: # runs built container locally
	docker run --rm -it -p 8080:8080 --name example-service ghcr.io/example/service:latest
