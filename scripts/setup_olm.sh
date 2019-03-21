#!/bin/bash

set -x +e

QUAY_USER="cmoulliard"
TMP_ROOT="$(pwd)/tmp"
CONSOLE="$(greadlink -f $TMP_ROOT/console)"
DEVOPS_CONSOLE_OPERATOR="$(greadlink -f $GOPATH/src/github.com/redhat-developer/devopsconsole-operator)"

# Step 1. Clone anf build the New Openshift Console
if [ ! -f $CONSOLE/bin/bridge ]; then
    rm -rvf $CONSOLE
    git clone --depth=1 https://github.com/talamer/console $CONSOLE

    cd $CONSOLE
    ./build.sh
fi

# Step 2. Configure the console and launch it locally

cd $CONSOLE
source ./contrib/oc-environment.sh
./bin/bridge &
BRIDGE_PID=$!

#sleep 5
#
#CONSOLE_STATUS=$(curl --silent -LI http://localhost:9000 | head -n 1 | cut -d ' ' -f 2)
#
#if [[ $CONSOLE_STATUS -neq 200 ]]; then
#    kill $BRIDGE_PID
#    echo "ERROR"oc
#    exit 1
#fi

# Step 3. Create a project and install the OLM
oc apply -f https://raw.githubusercontent.com/operator-framework/operator-lifecycle-manager/master/deploy/upstream/quickstart/olm.yaml

# Step 4. Git clone the devops operator and build it
git clone --depth=1 https://github.com/redhat-developer/devopsconsole-operator $DEVOPS_CONSOLE_OPERATOR
cd $DEVOPS_CONSOLE_OPERATOR
make build


# Step 5. Push it to Quay
operator-sdk build quay.io/$pmacik/devopsconsole-operator
docker push quay.io/$pmacik/devopsconsole-operator
