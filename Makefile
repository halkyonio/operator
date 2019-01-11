VERSION     ?= unset
GITCOMMIT    := $(shell git rev-parse --short HEAD 2>/dev/null)
PROJECT_NAME := component-operator
BIN_DIR      := ./build/_output/bin
REPO_PATH    := github.com/snowdrop/$(PROJECT_NAME)
BUILD_PATH   := $(REPO_PATH)/cmd/manager
BUILD_FLAGS  := -ldflags="-w -X main.Version=$(VERSION) -X main.GitCommit=$(GITCOMMIT)"

GO           ?= go
GOFMT        ?= $(GO)fmt
GOFILES      := $(shell find . -type f -name '*.go' -not -path "./vendor/*")

# go get -u github.com/shurcooL/vfsgen/cmd/vfsgendev
VFSGENDEV   := $(GOPATH)/bin/vfsgendev
PREFIX      ?= $(shell pwd)

.PHONY: clean
clean:
	@echo "> Remove build dir"
	@rm -rf ./build

.PHONY: build
build: clean
	@echo "> Build go application"
	go build ${BUILD_FLAGS} -o ${BIN_DIR}/${PROJECT_NAME} ${BUILD_PATH}

.PHONY: build-linux
build-linux: clean
	@echo "> Build go application for linux os"
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ${BUILD_FLAGS} -o ${BIN_DIR}/${PROJECT_NAME} ${BUILD_PATH}

.PHONY: cross
cross: clean
	@echo "> Build go application cross os"
	gox -osarch="darwin/amd64 linux/amd64" -output="${BIN_DIR}/bin/{{.OS}}-{{.Arch}}/${PROJECT_NAME}" $(BUILD_FLAGS) ${BUILD_PATH}

.PHONY: assets
assets: $(VFSGENDEV)
	@echo ">> writing assets"
	cd $(PREFIX)/pkg/util/template && go generate

$(VFSGENDEV):
	cd $(PREFIX)/vendor/github.com/shurcooL/vfsgen/ && go install ./cmd/vfsgendev/...

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
	gometalinter ./... --vendor

.PHONY: test-e2e
test-e2e:
	go test -v $(REPO_PATH)/e2e -ginkgo.v

.PHONY: unit-test
unit-test:
	go test ./test/main_test.go -root=$(PREFIX) -kubeconfig=$$HOME/.kube/config -namespacedMan deploy/namespace-init.yaml -globalMan deploy/crd.yaml -v -parallel=2

.PHONY: prepare-release
prepare-release: cross
	./scripts/prepare_release.sh

.PHONY: upload
upload: prepare-release
	./scripts/upload_assets.sh