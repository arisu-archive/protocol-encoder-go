.DEFAULT_GOAL := help

GO ?= go
CMAKE ?= cmake
UNICORN_BUILD_DIR := $(CURDIR)/cmake-build-unicorn
UNICORN_ENV := CGO_ENABLED=1 CGO_CFLAGS="-I$(CURDIR)/unicorn/include" CGO_LDFLAGS="-L$(UNICORN_BUILD_DIR)" LD_LIBRARY_PATH="$(UNICORN_BUILD_DIR):$${LD_LIBRARY_PATH}" PATH="$(UNICORN_BUILD_DIR):$${PATH}"

ARONA_INPUT_DIR := libraries/com.nexon.bluearchive
PLANA_INPUT_DIR := libraries/com.YostarJP.BlueArchive

.PHONY: help test test-race test-tools vet fmt unicorn-patch unicorn-build generate-arona generate-plana check-arona check-plana check-generated verify

help: ## Show available targets
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z_-]+:.*## / {printf "%-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

test: ## Test the runtime library
	$(GO) test ./... -count=1

test-race: ## Test the runtime library with the race detector
	$(GO) test -race ./... -count=1

vet: ## Vet the runtime library
	$(GO) vet ./...

fmt: ## Format root and tools Go packages
	$(GO) fmt ./...
	$(GO) -C tools fmt ./...

unicorn-patch: ## Apply repository patches required by the tools module
	sh ./scripts/apply_patch.sh

unicorn-build: unicorn-patch ## Build Unicorn for the tools module
	$(CMAKE) -S unicorn -B $(UNICORN_BUILD_DIR) -G Ninja -DCMAKE_BUILD_TYPE=Release -DUNICORN_BUILD_SHARED=ON -DUNICORN_ARCH=aarch64
	$(CMAKE) --build $(UNICORN_BUILD_DIR) --config Release --parallel

test-tools: unicorn-build ## Test the tools module
	$(UNICORN_ENV) $(GO) -C tools test ./... -count=1

generate-arona: unicorn-build ## Generate the Arona encoder table from existing inputs
	$(UNICORN_ENV) $(GO) -C tools run ./generator --package arona --library ../$(ARONA_INPUT_DIR)/libil2cpp.so --offset ../$(ARONA_INPUT_DIR)/offset.txt --version "$$(cat $(ARONA_INPUT_DIR)/version.txt)" --output ../pkg/encoder/arona/table_gen.go

generate-plana: unicorn-build ## Generate the Plana encoder table from existing inputs
	$(UNICORN_ENV) $(GO) -C tools run ./generator --package plana --library ../$(PLANA_INPUT_DIR)/libil2cpp.so --offset ../$(PLANA_INPUT_DIR)/offset.txt --version "$$(cat $(PLANA_INPUT_DIR)/version.txt)" --output ../pkg/encoder/plana/table_gen.go

check-arona: unicorn-build ## Verify the generated Arona encoder table
	$(UNICORN_ENV) $(GO) -C tools run ./generator --package arona --library ../$(ARONA_INPUT_DIR)/libil2cpp.so --offset ../$(ARONA_INPUT_DIR)/offset.txt --version "$$(cat $(ARONA_INPUT_DIR)/version.txt)" --output ../pkg/encoder/arona/table_gen.go --check

check-plana: unicorn-build ## Verify the generated Plana encoder table
	$(UNICORN_ENV) $(GO) -C tools run ./generator --package plana --library ../$(PLANA_INPUT_DIR)/libil2cpp.so --offset ../$(PLANA_INPUT_DIR)/offset.txt --version "$$(cat $(PLANA_INPUT_DIR)/version.txt)" --output ../pkg/encoder/plana/table_gen.go --check

check-generated: check-arona check-plana ## Verify both generated tables

verify: test test-race vet test-tools ## Run root and tools verification
