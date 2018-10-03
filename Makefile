VERSION     ?= 0.0.1
GITCOMMIT    := $(shell git rev-parse --short HEAD 2>/dev/null)
PROJECT_NAME := spring-boot-operator
BIN_DIR      := ./tmp/_output/bin
REPO_PATH    := github.com/snowdrop/$(PROJECT_NAME)
BUILD_PATH   := $(REPO_PATH)/cmd/sb
BUILD_FLAGS := -ldflags="-w -X $(PROJECT)/cmd.GITCOMMIT=$(GITCOMMIT) -X $(PROJECT_NAME)/cmd.VERSION=$(VERSION)"
GO          ?= go

clean:
	@echo "> Remove dist dir"
	@rm -rf ./dist

build:
	@echo "> Build go application"
	go build ${BUILD_FLAGS} -o ${BIN_DIR}/sd ${BUILD_PATH}

cross: clean
	gox -osarch="darwin/amd64 linux/amd64" -output="${BIN_DIR}/bin/{{.OS}}-{{.Arch}}/sd" $(BUILD_FLAGS)