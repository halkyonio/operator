#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

TMP_DIR=$(mktemp -d)
CURRENT_DIR=$PWD
KIND=component
API_VERSION=v1alpha2
GEN_FILE=zz_generated.deepcopy.go
GEN_FILE_PATH=${KIND}/${API_VERSION}/${GEN_FILE}

cleanup() {
  rm -rf "${TMP_DIR}"
}
trap "cleanup" EXIT SIGINT

cleanup

# The following solution for making code generation work with go modules is
# borrowed and modified from https://github.com/heptio/contour/pull/1010.
export GO111MODULE=on
VERSION=$(go list -m all | grep k8s.io/code-generator | rev | cut -d"-" -f1 | cut -d" " -f1 | rev)
git clone https://github.com/kubernetes/code-generator.git ${TMP_DIR}
(cd ${TMP_DIR} && git reset --hard ${VERSION} && go mod init)

# Usage: generate-groups.sh <generators> <output-package> <apis-package> <groups-versions> ...
#
#   <generators>        the generators comma separated to run (deepcopy,defaulter,client,lister,informer) or "all".
#   <output-package>    the output package name (e.g. github.com/example/project/pkg/generated).
#   <apis-package>      the external types dir (e.g. github.com/example/api or github.com/example/project/pkg/apis).
#   <groups-versions>   the groups and their versions in the format "groupA:v1,v2 groupB:v1 groupC:v2", relative
#                       to <api-package>.
#   ...                 arbitrary flags passed to all generator binaries.
#
#
# Examples:
#   generate-groups.sh all             github.com/example/project/pkg/client github.com/example/project/pkg/apis "foo:v1 bar:v1alpha1,v1beta1"
#   generate-groups.sh deepcopy,client github.com/example/project/pkg/client github.com/example/project/pkg/apis "foo:v1 bar:v1alpha1,v1beta1"


${TMP_DIR}/generate-groups.sh deepcopy,client \
  github.com/snowdrop/component-operator/pkg/apis \
  github.com/snowdrop/component-operator/pkg/apis \
  ${KIND}:${API_VERSION} \
  --go-header-file ${TMP_DIR}/hack/boilerplate.go.txt

# NOT NEEDED IF WE INSTALL THE PROJECT UNDER GOPATH AND THAT WE SET GO111MODULE=on
# export GO111MODULE=on
# cp $GOPATH/src/github.com/snowdrop/component-operator/pkg/apis/${GEN_FILE_PATH} $CURRENT_DIR/pkg/apis/${GEN_FILE_PATH}
