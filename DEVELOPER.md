# Halkyon Developer information

Table of Contents
=================
  * [How to package the operator for OperatorHub](#how-to-package-the-operator-for-operatorhub)
     * [Build the docker image of the Operator](#build-the-docker-image-of-the-operator)
     * [Fork the community operators and prepare the OLM bundle](#fork-the-community-operators-and-prepare-the-olm-bundle)
     * [How to deploy the bundle as Quay application](#how-to-deploy-the-bundle-as-quay-application)
     * [How to deploy the Operator on the OLM registry of OCP4](#how-to-deploy-the-operator-on-the-olm-registry-of-ocp4)

## How to package the operator for OperatorHub

The following section explains the different steps that we must follow in order to do the work to deploy this
operator on `operatorhub.io`

- Create a Container image published next on Quay.io
- Create the OLM Bundle definition containing the CRDs, ClusterServiceVersion & Package resource
- Push using the `operator-courier` tool the bundle or application on Quay.io
- Prepare and submit a PR containing the bundle info using the forked project - github.com:operator-framework/community-operators.git

### Build the docker image of the Operator

The following steps are executed byt the `.circleci/config.yml` job's file and can also be used
locally to build your own docker image. 

```bash
VERSION:v0.1.x
docker build -t operator:${VERSION} -f build/Dockerfile .
TAG_ID=$(docker images -q operator:${VERSION})
docker tag ${TAG_ID} quay.io/halkyonio/operator:${VERSION}
docker tag ${TAG_ID} quay.io/halkyonio/operator:latest
docker login quay.io -u="${QUAY_ROBOT_USER}" -p="${QUAY_ROBOT_TOKEN}"
docker push quay.io/halkyonio/operator:${VERSION}
docker push quay.io/halkyonio/operator:latest
```

### Fork the community operators and prepare the OLM bundle

- Git clone locally the forked community-operators project - https://github.com/halkyionis/community-operators
- Next create under `upstream-community-operators` or `community-operators` or both folders a project having the name of the operator
with the following resources that you can find as example under the operator project - `deploy/olm-catalog/bundle`.

```bash
upstream-community-operators
  halkyon
    capability.v1beta1.crd.yaml
    component.v1beta1.crd.yaml
    link_v1beta1.crd.yaml
    halkyon.package.yaml
    halkyon.v0.1.3.clusterserviceversion.yaml
```
- Submit the PR when the following step has been accomplished

**Remarks**

You can find more info about the definition of the `ClusterServiceVersion` resource and package file [here](https://github.com/snowdrop/community-operators/blob/master/docs/contributing.md#package-your-operator)
The `bundle` format describing the resources, versions part of the operator are described [here](https://github.com/snowdrop/community-operators/blob/master/docs/contributing.md#bundle-format)

The ClusterServiceVersion yaml resource can be either validated by the `operator-courier` tool or only at this [address](https://operatorhub.io/preview). The online tool will not 
only check the syntax, mandatory fields but will also display graphically the description, logo, ...

### How to deploy the bundle as Quay application

One of the requirement to let you to use an Operator with an OLM registry is to publish first it on `quay.io` the bundle information created previously
For that purpose, we will use the `operator-courier` tool which can validate or publish the bundle on quay.io

Install first the [tool](https://github.com/operator-framework/operator-courier) `operator-courier`.

    pip3 install operator-courier
    
Verify your operator's bundle using the tool.

    export BUNDLE_DIR="deploy/olm-catalog/bundle"
    operator-courier verify $BUNDLE_DIR  
    
**Remark** The `BUNDLE_DIR` must point to the directory containing the bundle to be tested and not yet published on `operatorhub.io`    

Next, get from `quay.io` an `Authentication token` using your `quay's username` OR `robot username` and password to access your namespace.

Next, execute the following `curl` request to get a token (e.g `basic Y2gwMDdtK...A="`).

    export QUAY_USER="QUAY USER"
    export QUAY_PWD="QUAY PASSWORD"    
    export AUTH_TOKEN=$(curl -X POST -H "Content-Type: application/json" -d '{"user":{"username":"'"$QUAY_USER"'","password":"'"$QUAY_PWD"'"}}' https://quay.io/cnr/api/v1/users/login | jq -r '.token')
    
Push finally the bundle on quay as an `application`.

    export QUAY_ORG="quay_organization (e.g halkyionio)"
    export APP_REPOSITORY="halkyon"
    export RELEASE="0.1.3"
    operator-courier push $BUNDLE_DIR $QUAY_ORG $APP_REPOSITORY $RELEASE "$AUTH_TOKEN"
    
**REMARK**: Use as version for the `RELEASE`, the tag id published under the github repository but without the letter `v`. Example : `0.1.3` and not `v0.1.1` !    
    
The following bash script can also be used to publish the bundle to be tested on quay

    ./scripts/release_bundle_operator.sh $QUAY_USER $QUAY_TOKEN $RELEASE    
    
**Warning**: The name of the `application repository` must match the name of the operator created under the folder `upstream-community-operators` or `community-operators`. The version, of course, will match the one defined within the `CSV` yaml resource or bundle package
    
### How to deploy the Operator on the OLM registry of OCP4

For local testing purposes, the bundle created previously can be tested using an ocp4 cluster. For that purpose, we will deploy different resources in order to let the OLM registry to fetch
from quay.io the bundle, install it and next to create a subscription in order to deploy the operator.

So log on first to an ocp4 cluster with a user having the `cluster-admin` role.
Next, deploy the `OperatorSource` in order to add a new registry that OLM will use to fetch from `Quay.io/app` the bundle of the operator.

    oc apply -f deploy/olm-catalog/ocp4/operator-source.yaml

Now, using the ocp console, subscribe to the `operator` coming from the Quay.io registry by clicking on the button `install` of the `Component operator` that you can select from the screen
`operatorhub`. The link to access this screen is `https://console-openshift-console.apps.snowdrop.devcluster.openshift.com/operatorhub`

![select operator](img/select-operator-hub.png)
![install operator](img/install-operator.png)
![subscribe operator](img/create-subscription.png)

Wait a few moments and check if the pod of the operator has been created under the `openshift-operators` namespace.

    oc get -n openshift-operators pods
    NAME                                  READY     STATUS    RESTARTS   AGE
    halkyon-operator-85fcbdf6fc-r4fmf   1/1       Running   0          9m
    
To clean-up , execute the following commands

    oc delete -n openshift-operators subscriptions/halkyon
    oc delete -n openshift-marketplace operatorsource/halkyon-operators
    oc delete crd/components.halkyon.io,crd/links.halkyon.io,crd/capabilities.halkyon.io
    oc delete -n openshift-operators ClusterServiceVersion/halkyon.v0.1.3
    oc delete -n openshift-marketplace CatalogSourceConfig/halkyon-operators
    oc delete -n openshift-operators deployment/halkyon-operator
