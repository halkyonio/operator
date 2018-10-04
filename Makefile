VERSION     ?= 0.0.1
GITCOMMIT    := $(shell git rev-parse --short HEAD 2>/dev/null)
PROJECT_NAME := component-operator
BIN_DIR      := ./tmp/_output/bin
REPO_PATH    := github.com/snowdrop/$(PROJECT_NAME)
BUILD_PATH   := $(REPO_PATH)/cmd/sd
BUILD_FLAGS  := -ldflags="-w -X $(PROJECT)/cmd.GITCOMMIT=$(GITCOMMIT) -X $(PROJECT_NAME)/cmd.VERSION=$(VERSION)"

GO           ?= go
GOFMT        ?= $(GO)fmt
GOFILES      := $(shell find . -type f -name '*.go' -not -path "./vendor/*")

# go get -u github.com/shurcooL/vfsgen/cmd/vfsgendev
VFSGENDEV   := $(GOPATH)/bin/vfsgendev
PREFIX      ?= $(shell pwd)

clean:
	@echo "> Remove dist dir"
	@rm -rf ./dist

build:
	@echo "> Build go application"
	go build ${BUILD_FLAGS} -o ${BIN_DIR}/sd ${BUILD_PATH}

cross: clean
	gox -osarch="darwin/amd64 linux/amd64" -output="${BIN_DIR}/bin/{{.OS}}-{{.Arch}}/sd" $(BUILD_FLAGS)

assets: $(VFSGENDEV)
	@echo ">> writing assets"
	cd $(PREFIX)/pkg/util/template && go generate

$(VFSGENDEV):
	cd $(PREFIX)/vendor/github.com/shurcooL/vfsgen/ && go install ./cmd/vfsgendev/...

format:
	@echo ">> checking code style"
	@fmtRes=$$($(GOFMT) -d $$(find . -path ./vendor -prune -o -name '*.go' -print)); \
	if [ -n "$${fmtRes}" ]; then \
		echo "gofmt checking failed!"; echo "$${fmtRes}"; echo; \
		exit 1; \
	fi