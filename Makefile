VERSION     ?= unset
GITCOMMIT    := $(shell git rev-parse --short HEAD 2>/dev/null)
PROJECT_NAME := operator
BIN_DIR      := ./build/_output/bin
REPO_PATH    := halkyon.io/$(PROJECT_NAME)
BUILD_PATH   := $(REPO_PATH)/cmd/manager
BUILD_FLAGS  := -ldflags="-w -X main.Version=$(VERSION) -X main.GitCommit=$(GITCOMMIT)"
PKGS         := $(shell go list  ./... | grep -v $(PROJECT)/vendor)

GO           ?= go
GOFMT        ?= $(GO)fmt
GOFILES      := $(shell find . -type f -name '*.go' -not -path "./vendor/*")

PREFIX      ?= $(shell pwd)

default: build

.PHONY: clean
clean:
	@echo "> Remove build dir"
	@rm -rf ./build/_output

.PHONY: build
build: clean
	@echo "> Build go application"
	GO111MODULE=on go build ${BUILD_FLAGS} -o ${BIN_DIR}/halkyon-${PROJECT_NAME} ${BUILD_PATH}

.PHONY: build-linux
build-linux: clean
	@echo "> Build go application for linux os"
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ${BUILD_FLAGS} -o ${BIN_DIR}/halkyon-${PROJECT_NAME} ${BUILD_PATH}

.PHONY: cross
cross: clean
	@echo "> Build go application cross os"
	gox -osarch="darwin/amd64 linux/amd64" -output="${BIN_DIR}/{{.OS}}-{{.Arch}}/halkyon-${PROJECT_NAME}" $(BUILD_FLAGS) ${BUILD_PATH}

.PHONY: generate-api
generate-api:
	@echo "> Updating generated code"
	./scripts/update-gen.sh

.PHONY: format
format:
	@echo ">> checking code style"
	@fmtRes=$$($(GOFMT) -d $$(find . -path ./vendor -prune -o -name '*.go' -print)); \
	if [ -n "$${fmtRes}" ]; then \
		echo "gofmt checking failed!"; echo "$${fmtRes}"; echo; \
		exit 1; \
	fi

.PHONY: lint
lint:
	golint $(PKGS)

# Tests to be executed within k8s cluster
.PHONY: test-e2e
test-e2e:
	go test -v $(REPO_PATH)/e2e -ginkgo.v

# Tests using operator deployed in a cluster
.PHONY: unit-test
unit-test:
	go test ./test/main_test.go -root=$(PREFIX) -kubeconfig=$$HOME/.kube/config -namespacedMan deploy/namespace-init.yaml -globalMan deploy/crds/component.yaml

.PHONY: test
test:
	go test ./pkg/...

.PHONY: prepare-release
prepare-release: cross
	./scripts/prepare_release.sh

.PHONY: upload
upload: prepare-release
	./scripts/upload_assets_test.sh

dep:
	$(Q)dep ensure -v

dep-update:
	$(Q)dep ensure -update -v