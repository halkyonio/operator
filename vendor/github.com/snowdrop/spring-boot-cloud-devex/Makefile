VERSION     ?= 0.666.0

PROJECT     := github.com/snowdrop/spring-boot-cloud-devex
GITCOMMIT   := $(shell git rev-parse --short HEAD 2>/dev/null)
BUILD_FLAGS := -ldflags="-w -X $(PROJECT)/cmd.GITCOMMIT=$(GITCOMMIT) -X $(PROJECT)/cmd.VERSION=$(VERSION)"

GO          ?= go
GOFMT       ?= $(GO)fmt
GOFILES     := $(shell find . -type f -name '*.go' -not -path "./vendor/*")

# go get -u github.com/shurcooL/vfsgen/cmd/vfsgendev
VFSGENDEV   := $(GOPATH)/bin/vfsgendev
PREFIX      ?= $(shell pwd)

all: clean build

clean:
	@echo "> Remove dist dir"
	@rm -rf ./dist

build:
	@echo "> Build go application"
	go build ${BUILD_FLAGS} -o sb main.go

cross: clean
	gox -osarch="darwin/amd64 linux/amd64" -output="dist/bin/{{.OS}}-{{.Arch}}/sb" $(BUILD_FLAGS)

prepare-release: cross
	./scripts/prepare_release.sh

upload: prepare-release
	./scripts/upload_assets.sh

assets: $(VFSGENDEV)
	@echo ">> writing assets"
	cd $(PREFIX)/pkg/buildpack && go generate

$(VFSGENDEV):
	cd $(PREFIX)/vendor/github.com/shurcooL/vfsgen/ && go install ./cmd/vfsgendev/...

format:
	@echo ">> checking code style"
	@fmtRes=$$($(GOFMT) -d $$(find . -path ./vendor -prune -o -name '*.go' -print)); \
	if [ -n "$${fmtRes}" ]; then \
		echo "gofmt checking failed!"; echo "$${fmtRes}"; echo; \
		exit 1; \
	fi

version:
	@echo $(VERSION)