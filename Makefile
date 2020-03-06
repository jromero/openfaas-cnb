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

default: test build

help:
	@echo 'Management commands for openfaas-cnb:'
	@echo
	@echo 'Usage:'
	@echo '    make build           Compile the project.'
	@echo '    make get-deps        runs dep ensure, mostly used for ci.'
	
	@echo '    make clean           Clean the directory tree.'
	@echo

build: export GOFLAGS := $(GOFLAGS)
build: export GOOS := linux
build:
	@echo "> Building ${VERSION}..."
	go build -ldflags="$(LDFLAGS)" -o build/bin/build -a ./cmd/build
	go build -ldflags="$(LDFLAGS)" -o build/bin/detect -a ./cmd/detect
	cp buildpack.toml build/buildpack.toml
	cp package.toml build/package.toml

test: export GOFLAGS := $(GOFLAGS)
test:
	@echo "> Running tests..."
	go test -v ./...

test-e2e:
	@echo "> Running end-to-end tests..."
	go test -tags e2e -v ./test_e2e/...

package-image:
	@echo "> Packaging as image..."
	cd build; pack package-buildpack jar03/openfass-cnb:latest -p package.toml

package-tgz:
	@echo "> Packaging as tgz..."
	@cd build; tar cvzf openfaas-cnb-$(VERSION).tgz buildpack.toml bin/

clean:
	@test ! -e build || rm -rf build

.PHONY: clean test test-e2e