GO=go
CI ?= false
GOTESTFLAGS ?= -race -timeout 120s
GOTESTFLAGSNORACE = -timeout 120s
GOFLAGS=
UNAME := $(shell uname)
CODESIGN_IDENTITY ?= Developer ID Application: Unstable Build, LLC. (YYZRWD888J)

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

# Each release target selects its Bluectl project and destination from the
# local, gitignored deploy/bluectl matrix; provision it before publishing.
BLUECTL_HOST_OS := $(shell uname | awk '{print tolower($$0)}')
BLUECTL_HOST_ARCH := $(shell uname -m | sed -e 's/x86_64/amd64/' -e 's/aarch64/arm64/')
BLUECTL_CONFIG_ROOT := $(abspath deploy/bluectl)
BLUECTL_CONFIG = $(BLUECTL_CONFIG_ROOT)/$(1)/$(2)

.PHONY: debug clean test coverage generate \
	format license assert_license license_replace \
	runectl runectl-pkg runectl-sign \
	runectl-prod-dist runectl-staging-dist \
	runectl-prod-dist-linux-amd64 runectl-staging-dist-linux-amd64 \
	runectl-prod-dist-linux-arm64 runectl-staging-dist-linux-arm64 \
	runectl-prod-dist-darwin-amd64 runectl-staging-dist-darwin-amd64 \
	runectl-prod-dist-darwin-arm64 runectl-staging-dist-darwin-arm64 \
	runectl-prod-dist-all \
	dist-tar-with-src runectl-dist-version-test

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
	@$(MAKE) -C cmd/runectl clean

$(BIN):
	@mkdir $(BIN)

$(EXAMPLES): $(EXAMPLESRC) $(LIBSRC) $(BIN)
	@cd $(patsubst bin/example_%,examples/%,$@) && $(CGO_ENABLED) $(GO) build $(GOFLAGS) -o ../../$@

$(EXECS): $(EXECSRC) $(LIBSRC) $(BIN)
	@cd $(patsubst bin/%,cmd/%,$@) && $(CGO_ENABLED) $(GO) build $(GOFLAGS) -o ../../$@

runectl: $(BIN)/runectl

runectl-pkg:
	@$(MAKE) -C cmd/runectl pkg

runectl-sign:
	@$(MAKE) -C cmd/runectl sign

# Host-native release targets derive their artifact platform from the host.
# Set RUNECTL_VERSION explicitly (e.g. RUNECTL_VERSION=v1.1.3 make runectl-prod-dist).
# It is independent of the SDK source tag, which remains in the upload metadata.
runectl-prod-dist: clean
	@BLUECTL_CONFIG_DIR=$(call BLUECTL_CONFIG,prod,$(BLUECTL_HOST_OS)-$(BLUECTL_HOST_ARCH)) $(MAKE) -C cmd/runectl dist

runectl-staging-dist: clean
	@BLUECTL_CONFIG_DIR=$(call BLUECTL_CONFIG,staging,$(BLUECTL_HOST_OS)-$(BLUECTL_HOST_ARCH)) $(MAKE) -C cmd/runectl dist

# Cross-platform release targets select the destination config for the
# artifact's target platform rather than the host that builds it.
runectl-prod-dist-linux-amd64: clean
	@BLUECTL_CONFIG_DIR=$(call BLUECTL_CONFIG,prod,linux-amd64) $(MAKE) -C cmd/runectl dist-linux-amd64

runectl-staging-dist-linux-amd64: clean
	@BLUECTL_CONFIG_DIR=$(call BLUECTL_CONFIG,staging,linux-amd64) $(MAKE) -C cmd/runectl dist-linux-amd64

runectl-prod-dist-linux-arm64: clean
	@BLUECTL_CONFIG_DIR=$(call BLUECTL_CONFIG,prod,linux-arm64) $(MAKE) -C cmd/runectl dist-linux-arm64

runectl-staging-dist-linux-arm64: clean
	@BLUECTL_CONFIG_DIR=$(call BLUECTL_CONFIG,staging,linux-arm64) $(MAKE) -C cmd/runectl dist-linux-arm64

runectl-prod-dist-darwin-amd64: clean
	@BLUECTL_CONFIG_DIR=$(call BLUECTL_CONFIG,prod,darwin-amd64) $(MAKE) -C cmd/runectl dist-darwin-amd64

runectl-staging-dist-darwin-amd64: clean
	@BLUECTL_CONFIG_DIR=$(call BLUECTL_CONFIG,staging,darwin-amd64) $(MAKE) -C cmd/runectl dist-darwin-amd64

runectl-prod-dist-darwin-arm64: clean
	@BLUECTL_CONFIG_DIR=$(call BLUECTL_CONFIG,prod,darwin-arm64) $(MAKE) -C cmd/runectl dist-darwin-arm64

runectl-staging-dist-darwin-arm64: clean
	@BLUECTL_CONFIG_DIR=$(call BLUECTL_CONFIG,staging,darwin-arm64) $(MAKE) -C cmd/runectl dist-darwin-arm64

# Each platform target cleans its own build outputs. Run them in separate
# sub-makes so a parallel make cannot clean another platform's in-flight build.
runectl-prod-dist-all:
	@$(MAKE) runectl-prod-dist-linux-amd64
	@$(MAKE) runectl-prod-dist-linux-arm64
	@$(MAKE) runectl-prod-dist-darwin-amd64
	@$(MAKE) runectl-prod-dist-darwin-arm64

# Drive the real runectl publisher with an artifact that leaks a .go file and
# assert the final pre-publish gate aborts before Bluectl can be invoked.
dist-tar-with-src:
	@./cmd/dist-with-src-test.sh tar

runectl-dist-version-test:
	@bash ./cmd/runectl/dist-version-test.sh
