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
oc login -u system:admin

export OPENSHIFT_API="$(oc status | grep "on server http" | sed -e 's,.*on server \(http.*\),\1,g')"

oc process -f $CONSOLE/examples/console-oauth-client.yaml | oc apply -f -
oc get oauthclient console-oauth-client -o jsonpath='{.secret}' > $CONSOLE/examples/console-client-secret
oc get secrets -n default --field-selector type=kubernetes.io/service-account-token -o json | jq '.items[0].data."ca.crt"' -r | python -m base64 -d > $CONSOLE/examples/ca.crt
oc adm policy add-cluster-role-to-user cluster-admin admin

cd $CONSOLE
$CONSOLE/examples/run-bridge.sh &
BRIDGE_PID=$!

sleep 5

CONSOLE_STATUS=$(curl --silent -LI http://localhost:9000 | head -n 1 | cut -d ' ' -f 2)

if [ $CONSOLE_STATUS -neq 200 ]; then
    kill $BRIDGE_PID
    echo "ERROR"
    exit 1
fi

# Step 3. Create a project and install the OLM
oc project myproject
oc create -f https://raw.githubusercontent.com/operator-framework/operator-lifecycle-manager/master/deploy/upstream/quickstart/olm.yaml

# Step 4. Git clone the devops operator and build it
git clone --depth=1 https://github.com/redhat-developer/devopsconsole-operator $DEVOPS_CONSOLE_OPERATOR
cd $DEVOPS_CONSOLE_OPERATOR
make build


# Step 5. Push it to Quay
operator-sdk build quay.io/$pmacik/devopsconsole-operator
docker push quay.io/$pmacik/devopsconsole-operator
