GO=go
CI ?= false
GOTESTFLAGS ?= -race -timeout 120s
GOTESTFLAGSNORACE = -timeout 120s
GOFLAGS=
UNAME := $(shell uname)

BIN=bin
TARGET=target
LIBSRC=$(wildcard *.go) $(wildcard **/*.go) $(wildcard **/**/*.go) $(wildcard **/**/**/*.go)
EXAMPLESRC=$(wildcard examples/*/*.go)
EXAMPLEDIRS=$(sort $(dir $(EXAMPLESRC)))
EXAMPLES=$(patsubst examples/%/,$(BIN)/example_%,$(EXAMPLEDIRS))
EXECSRC=$(wildcard cmd/**/*.go) $(wildcard cmd/**/**/*.go)
EXECDIRS=$(sort $(dir $(EXECSRC)))
EXECS=$(patsubst cmd/%/,$(BIN)/%,$(EXECDIRS))
GOMOCKS=$(wildcard **/**/*_gomock.go) $(wildcard **/*_gomock.go)
RELEASE_FILES=$(wildcard release/*)

.PHONY: debug clean test coverage generate \
	format license assert_license license_replace

default: .git/hooks/pre-commit $(EXAMPLES) $(EXECS)

debug: GOFLAGS=-race
debug: $(EXAMPLES) $(EXECS)

.git/hooks/pre-commit: .pre-commit-config.yaml
	@ pre-commit install

test: CI=$(CI)
test:
	@ go test ./.../... $(GOTESTFLAGS)

test: CI=$(CI)
test-no-race:
	@ go test ./.../... $(GOTESTFLAGSNORACE)

coverage: $(BIN)
	@ go test ./.../... -coverprofile $(BIN)/coverage
	@ go tool cover -html=$(BIN)/coverage

generate:
	@ rm -rf **/**/*rpc*/*.pb.go
	@ go generate ./...

license:
	@ bluectl license LICENSE `find . -name \*.go | grep -v gomock | grep -v .pb.go | xargs`

license_replace:
	@ bluectl license -f LICENSE `find . -name \*.go | grep -v gomock | grep -v .pb.go | xargs`

assert_license:
	@ bluectl license -d LICENSE `find . -name \*.go | grep -v gomock | grep -v .pb.go | xargs`

format:
	@ go fmt ./.../...

cross-compile:
	@ . ./test_crosscompile.sh

lint:
	@ golangci-lint run --timeout=600s

clean:
	@rm -rf $(BIN) $(TARGET)

$(BIN):
	@mkdir $(BIN)

$(EXAMPLES): $(EXAMPLESRC) $(LIBSRC) $(BIN)
	@cd $(patsubst bin/example_%,examples/%,$@) && $(CGO_ENABLED) $(GO) build $(GOFLAGS) -o ../../$@

$(EXECS): $(EXECSRC) $(LIBSRC) $(BIN)
	@cd $(patsubst bin/%,cmd/%,$@) && $(CGO_ENABLED) $(GO) build $(GOFLAGS) -o ../../$@
