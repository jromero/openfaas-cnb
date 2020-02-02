.PHONY: build build-alpine clean test help default

VERSION:=$(shell grep "version" buildpack.toml | sed -E 's/.*"(.+)"$$/\1/')
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_DIRTY=$(shell test -n "`git status --porcelain`" && echo "+CHANGES" || true)
BUILD_DATE=$(shell date '+%Y-%m-%d-%H:%M:%S')

LDFLAGS=-s -w
LDFLAGS+=-X 'github.com/jromero/openfaas-cnb/pkg/version.GitCommit=${GIT_COMMIT}${GIT_DIRTY}'
LDFLAGS+=-X 'github.com/jromero/openfaas-cnb/pkg/version.BuildDate=${BUILD_DATE}'
LDFLAGS+=-X 'github.com/jromero/openfaas-cnb/pkg/version.Version=${VERSION}'

GOFLAGS?=-mod=vendor
GOOS?=linux

default: test

help:
	@echo 'Management commands for openfaas-cnb:'
	@echo
	@echo 'Usage:'
	@echo '    make build           Compile the project.'
	@echo '    make get-deps        runs dep ensure, mostly used for ci.'
	
	@echo '    make clean           Clean the directory tree.'
	@echo

build: export GOFLAGS := $(GOFLAGS)
build: export GOOS := $(GOOS)
build:
	@echo "> Building ${BIN_NAME} ${VERSION}..."
	go build -ldflags="$(LDFLAGS)" -o bin/build -a ./cmd/build
	go build -ldflags="$(LDFLAGS)" -o bin/detect -a ./cmd/detect

clean:
	@test ! -e bin/build || rm bin/build
	@test ! -e bin/detect || rm bin/detect

test:
	go test ./...

test-e2e:
	@echo "> Runing tests..."
	./test-e2e/test.sh

.PHONY: clean test test-e2e