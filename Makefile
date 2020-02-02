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
GOARCH?=amd64

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
build: export GOARCH := $(GOARCH)
build: build/
	@echo "> Building ${BIN_NAME} ${VERSION}..."
	go build -ldflags="$(LDFLAGS)" -o build/bin/build -a ./cmd/build
	go build -ldflags="$(LDFLAGS)" -o build/bin/detect -a ./cmd/detect
	cp buildpack.toml build/buildpack.toml

clean:
	@test ! -e build || rm -rf build

test:
	go test -v ./...

test-e2e:
	@echo "> Runing tests..."
	go test -tags e2e -v ./test_e2e/...

package:
	@echo "> Packaging..."
	@cd build; tar cvzf openfaas-cnb-$(VERSION).tgz buildpack.toml bin/

build/:
	mkdir -p build/

.PHONY: clean test test-e2e