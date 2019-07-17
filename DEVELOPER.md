# Developer information

Table of Contents
=================

  * [How To create the operator, crd](#how-to-create-the-operator-crd)
  * [How to package and install the Operator on Quay.io as an Application](#how-to-package-and-install-the-operator-on-quayio-as-an-application)
     * [How to deploy the Component Operator on OCP4](#how-to-deploy-the-component-operator-on-ocp4)
  * [Deprecated](#deprecated)
     * [How to install the Operator using OLM, Marketplace on okd 3.x](#how-to-install-the-operator-using-olm-marketplace-on-okd-3x)
        * [Step 1 - Install the new OpenShift console](#step-1---install-the-new-openshift-console)
        * [Step 2 - Install the OLM and Marketplace](#step-2---install-the-olm-and-marketplace)
        * [Step 3 - Install our Component Operator Source](#step-3---install-our-component-operator-source)
        * [Step 4 - Deploy the Component Operator using a Subscription](#step-4---deploy-the-component-operator-using-a-subscription)
     * [How to install okd 3.11 using the OCP resources](#how-to-install-okd-311-using-the-ocp-resources)
     * [Cleanup](#cleanup)


## How To create the operator, crd

Instructions followed to create the Component's CRD, operator using the `operator-sdk`'s kit

- Install the `operator-sdk` 
  ```bash
  brew install operator-sdk
  ```

- Execute this command within the `$GOPATH/github.com/$ORG/` folder in a terminal
  ```bash
  operator-sdk new component-operator --api-version=devexp.runtime.redhat.com/v1alpha2 --kind=Component --skip-git-init
  operator-sdk add api --api-version=devexp.runtime.redhat.com/v1alpha2 --kind=Component 
  ```
  using the following parameters 

  - Name of the folder to be created : `component-operator`
  - Api Group Name   : `component.k8s.io`
  - Api Version      : `v1alpha2`
  - Kind of Resource : `Component`

- Build and push the `component-operator` image to `quai.io`s registry
  ```bash
  $ operator-sdk build quay.io/snowdrop/component-operator
  $ docker push quay.io/snowdrop/component-operator
  ```
  
- Update the operator's manifest to use the built image name
  ```bash
  sed -i 's|REPLACE_IMAGE|quay.io/snowdrop/component-operator|g' deploy/operator.yaml
  ```
- Create a namespace `component-operator`

- Deploy the component-operator
  ```bash
  oc new-project component-operator
  oc create -f deploy/sa.yaml
  oc create -f deploy/cluster-rbac.yaml
  oc create -f deploy/crds/component_v1alpha2.yaml
  oc create -f deploy/operator.yaml
  ```

- By default, creating a custom resource triggers the `component-operator` to deploy a busybox pod
  ```bash
  oc create -f deploy/component/cr.yaml
  ```

- Verify that the busybox pod is created
  ```bash
  oc get pod -l app=busy-box
  NAME            READY     STATUS    RESTARTS   AGE
  busy-box   1/1       Running   0          50s
  ```

- Cleanup
  ```bash
  oc delete -f deploy/crds/component_v1alpha2.yaml
  oc delete -f deploy/operator.yaml
  oc delete -f deploy/cluster-rbac.yaml
  oc delete -f deploy/sa.yaml
  ```

- Start operator locally

  ```bash
  $ oc new-project my-spring-app
  $ OPERATOR_NAME=component-operator WATCH_NAMESPACE=my-spring-app KUBERNETES_CONFIG=/Users/dabou/.kube/config go run cmd/manager/main.go
  
  $ oc delete components,route,svc,is,pvc,dc --all=true && go run cmd/sd/sd.go create my-spring-boot
  OR
  $ oc apply -f deploy/component1.yml
  $ oc get all,pvc,components,dc
  
  oc delete components,route,svc,is,pvc,dc --all=true
  ```  

## How to package the operator for the OperatorHub

The following section explains the different steps that we must follow in order to do the work to deploy this
operator on `Operatorhub.io`

- Create a Container image published next on Quay.io
- Create the OLM Bundle definition containing the CRDs, ClusterServiceVersion & Package resource
- Push using the tool operator-courier the bundle or application on Quay.io
- Prepare and submit a PR containing the bundle info using the forked project - github.com:operator-framework/community-operators.git

### Build the docker image of the Operator

The following steps are defined within the .circleci config.yaml file 

```bash
docker build -t component-operator:${VERSION} -f build/Dockerfile .
TAG_ID=$(docker images -q component-operator:${VERSION})
docker tag ${TAG_ID} quay.io/snowdrop/component-operator:${VERSION}
docker tag ${TAG_ID} quay.io/snowdrop/component-operator:latest
docker login quay.io -u="${QUAY_ROBOT_USER}" -p="${QUAY_ROBOT_TOKEN}"
docker push quay.io/snowdrop/component-operator:${VERSION}
docker push quay.io/snowdrop/component-operator:latest
```

### Fork the community operators and prepare the OLM bundle

- Git clone locally the snowdrop community-operators project - https://github.com/snowdrop/community-operators
- Next create under `upstream-community-operators` or `community-operators` or both folders a project having the name of the operator
with the following resources that you can find as example under the operator project - `deploy/olm-catalog/bundle`.

```bash
upstream-community-operators
  <operator_name>
    capability.v1alpha2.crd.yaml
    component.v1alpha2.crd.yaml
    kubecomposer.package.yaml
    kubecomposer.v0.0.1.clusterserviceversion.yaml
    link_v1alpha2.crd.yaml
```

- Submit the PR when the following step has been accomplished

**Remarks**

You can find more info about the definition of the `ClusterServiceVersion` resource and package file [here](https://github.com/snowdrop/community-operators/blob/master/docs/contributing.md#package-your-operator)
The `bundle` format describing the resources, versions part of the operator are described [here](https://github.com/snowdrop/community-operators/blob/master/docs/contributing.md#bundle-format)

The ClusterServiceVersion yaml resource can be either validated by the `operator-courier` tool or only at this [address](https://operatorhub.io/preview). The online tool will not 
only check the syntax, mandatory fields but will also display graphically the description, logo, ...

### How to deploy the bundle as Quay application

One of the requirement to let you to use your Operator with an OLM registry is to publish first on quay.io the bundle information created previously
For that purpose, we will use the `operator-courier` tool which can validate or publish the bundle on quay.io

Install first the [tool](https://github.com/operator-framework/operator-courier) `operator-courier`.

    pip3 install operator-courier
    
Verify your operator's bundle using the tool.

    export BUNDLE_DIR="deploy/olm-catalog/bundle"
    operator-courier verify $BUNDLE_DIR  
    
**Remark** The `BUNDLE_DIR` must point to the directory containing the bundle to be tested and not yet published on operatorhub.io    

Next, get from `quay.io` an `Authentication token` using your quay's username OR robot username/pwd to access your namespace.

Next, execute the following `curl` request to get a token (e.g `basic Y2gwMDdtK...A="`).

    export QUAY_USER="QUAY USER"
    export QUAY_PWD="QUAY PASSWORD"    
    export AUTH_TOKEN=$(curl -X POST -H "Content-Type: application/json" -d '{"user":{"username":"'"$QUAY_USER"'","password":"'"$QUAY_PWD"'"}}' https://quay.io/cnr/api/v1/users/login | jq -r '.token')
    
Push finally the bundle on quay as an `application`.

    export QUAY_ORG="quay_organization (e.g ch007m)"
    export REPOSITORY="component"
    export RELEASE="0.10.0"
    operator-courier push $BUNDLE_DIR $QUAY_ORG $REPOSITORY $RELEASE "$AUTH_TOKEN"
    
**Warning**: The name of the repository must match the name of the operator created under the folder `upstream-community-operators` or `community-operators`. The version, of course, will match the one defined within the `CSV` yaml resource or bundle package
    

### How to deploy the Operator on the OLM registry of OCP4

For local testing purposes, the bundle created previously can be tested using an ocp4 cluster. For that purpose, we will deploy different resources in order to let the OLM registry to fetch
from quay.io the bundle, install it and next to create a subscription in order to deploy the operator.

So log on first to an ocp4 cluster with a user having the `cluster-admin` role.
Next, deploy the `OperatorSource` in order to install from `Quay.io/app` the bundle of the operator.

    oc apply -f deploy/olm-catalog/ocp4/operator-source.yaml

Now, using the ocp console, subscribe to the `operator` by clicking on the button `install` of the `Component operatror` that you can select from the screen
`operatorhub`.

![select operator](img/select-operator-hub.png)
![install operator](img/install-operator.png)
![subscribe operator](img/create-subscription.png)

Wait a few moments and check if the pod of the operator has been created under the `openshift-operators` namespace.

    oc get -n openshift-operators pods
    NAME                                  READY     STATUS    RESTARTS   AGE
    component-operator-85fcbdf6fc-r4fmf   1/1       Running   0          9m
    
To clean-up , execute the following commands

    oc delete -n openshift-operators subscriptions/component
    oc delete -n openshift-marketplace operatorsource/component-operator
    oc delete crd/components.component.k8s.io
    oc delete -n openshift-operators ClusterServiceVersion/component-operator.v0.10.0
    oc delete -n openshift-marketplace CatalogSourceConfig/installed-custom-openshift-operators 
    oc delete -n openshift-operators deployment/component-operator

## Deprecated

### How to install the Operator using OLM, Marketplace on okd 3.x

#### Step 1 - Install the new OpenShift console

- Launch an okd 3.11 cluster locally
- Log on with a user having the `cluster-admin` role
- Git clone the new OpenShift console (created for ocp4) and build it
```
git clone https://github.com/openshift/console.git && cd console
./build.sh
```

- Launch the proxy console locally 

```
source ./contrib/oc-environment.sh
./bin/bridge
```

#### Step 2 - Install the OLM and Marketplace 

From another terminal, install the *Operator Lifecycle Manager* using this command:
```bash
oc apply -f https://raw.githubusercontent.com/operator-framework/operator-lifecycle-manager/master/deploy/upstream/manifests/0.9.0/0000_50_olm_00-namespace.yaml
oc apply -f https://raw.githubusercontent.com/operator-framework/operator-lifecycle-manager/master/deploy/upstream/manifests/0.9.0/0000_50_olm_01-olm-operator.serviceaccount.yaml
oc apply -f https://raw.githubusercontent.com/operator-framework/operator-lifecycle-manager/master/deploy/upstream/manifests/0.9.0/0000_50_olm_03-clusterserviceversion.crd.yaml
oc apply -f https://raw.githubusercontent.com/operator-framework/operator-lifecycle-manager/master/deploy/upstream/manifests/0.9.0/0000_50_olm_04-installplan.crd.yaml
oc apply -f https://raw.githubusercontent.com/operator-framework/operator-lifecycle-manager/master/deploy/upstream/manifests/0.9.0/0000_50_olm_05-subscription.crd.yaml
oc apply -f https://raw.githubusercontent.com/operator-framework/operator-lifecycle-manager/master/deploy/upstream/manifests/0.9.0/0000_50_olm_06-catalogsource.crd.yaml
oc apply -f https://raw.githubusercontent.com/operator-framework/operator-lifecycle-manager/master/deploy/upstream/manifests/0.9.0/0000_50_olm_07-olm-operator.deployment.yaml
oc apply -f https://raw.githubusercontent.com/operator-framework/operator-lifecycle-manager/master/deploy/upstream/manifests/0.9.0/0000_50_olm_08-catalog-operator.deployment.yaml
oc apply -f https://raw.githubusercontent.com/operator-framework/operator-lifecycle-manager/master/deploy/upstream/manifests/0.9.0/0000_50_olm_09-aggregated.clusterrole.yaml
oc apply -f https://raw.githubusercontent.com/operator-framework/operator-lifecycle-manager/master/deploy/upstream/manifests/0.9.0/0000_50_olm_10-operatorgroup.crd.yaml
oc apply -f https://raw.githubusercontent.com/operator-framework/operator-lifecycle-manager/master/deploy/upstream/manifests/0.9.0/0000_50_olm_11-olm-operators.configmap.yaml
oc apply -f https://raw.githubusercontent.com/operator-framework/operator-lifecycle-manager/master/deploy/upstream/manifests/0.9.0/0000_50_olm_12-olm-operators.catalogsource.yaml
oc apply -f https://raw.githubusercontent.com/operator-framework/operator-lifecycle-manager/master/deploy/upstream/manifests/0.9.0/0000_50_olm_13-operatorgroup-default.yaml
oc apply -f https://raw.githubusercontent.com/operator-framework/operator-lifecycle-manager/master/deploy/upstream/manifests/0.9.0/0000_50_olm_14-packageserver.subscription.yaml
```

**Remark** the following command fails `oc create -f https://raw.githubusercontent.com/operator-framework/operator-lifecycle-manager/master/deploy/upstream/quickstart/olm.yaml` as some resources can't be installed
. This is the reason why all the resources are installed individually.

When installed, deploy the marketplace resources responsible to manage the different operators under the `Operatorhub`.
```bash
echo "Install marketplace"
oc create -f https://raw.githubusercontent.com/operator-framework/operator-marketplace/master/deploy/upstream/01_namespace.yaml
oc create -f https://raw.githubusercontent.com/operator-framework/operator-marketplace/master/deploy/upstream/02_catalogsourceconfig.crd.yaml
oc create -f https://raw.githubusercontent.com/operator-framework/operator-marketplace/master/deploy/upstream/03_operatorsource.crd.yaml
oc create -f https://raw.githubusercontent.com/operator-framework/operator-marketplace/master/deploy/upstream/04_service_account.yaml
oc create -f https://raw.githubusercontent.com/operator-framework/operator-marketplace/master/deploy/upstream/05_role.yaml
oc create -f https://raw.githubusercontent.com/operator-framework/operator-marketplace/master/deploy/upstream/06_role_binding.yaml
oc create -f https://raw.githubusercontent.com/operator-framework/operator-marketplace/master/deploy/upstream/07_operator.yaml
```

Add the `upstream OR community` operator source to fetch the data from the `quay/cnr` repo
```bash
oc create -f https://raw.githubusercontent.com/operator-framework/operator-marketplace/master/deploy/examples/upstream.operatorsource.cr.yaml -n marketplace
oc create -f https://raw.githubusercontent.com/operator-framework/operator-marketplace/master/deploy/examples/community.operatorsource.cr.yaml -n marketplace
```

#### Step 3 - Install our Component Operator Source

Install our `Component Operator` source too from `quay`
```bash
oc create -f deploy/olm-catalog/component-operator-source.yaml -n marketplace
```

After a few moment, check if the `OperatorSource` have been deployed successfully and that we collected their `package manifests`
```bash
oc get opsrc upstream-community-operators -o=custom-columns=NAME:.metadata.name,PACKAGES:.status.packages -n marketplace
NAME                           PACKAGES
upstream-community-operators   jaeger,prometheus,aws-service,etcd,mongodb-enterprise,redis-enterprise,federation,planetscale,strimzi-kafka-operator,cockroachdb,microcks,vault,percona,couchbase-enterprise,postgresql,oneagent

OR

oc get opsrc community-operators -o=custom-columns=NAME:.metadata.name,PACKAGES:.status.packages -n marketplace 
NAME                  PACKAGES
community-operators   etcd,prometheus,automationbroker,cockroachdb,hazelcast-enterprise,metering,jaeger,cluster-logging,node-problem-detector,templateservicebroker,postgresql,kiecloud-operator,planetscale,elasticsearch-operator,node-network-operator,microcks,descheduler,camel-k,oneagent,percona,strimzi-kafka-operator,federation

oc get opsrc component-operator -n marketplace
NAME                 TYPE          ENDPOINT              REGISTRY   DISPLAYNAME          PUBLISHER   STATUS      MESSAGE                                       AGE
component-operator   appregistry   https://quay.io/cnr   ch007m     Component Operator   Snowdrop    Succeeded   The object has been successfully reconciled   35s

oc get packagemanifests
NAME                     AGE
packageserver            5m
aws-service              5m
cockroachdb              5m
couchbase-enterprise     5m
etcd                     5m
federation               5m
jaeger                   5m
microcks                 5m
mongodb-enterprise       5m
oneagent                 5m
percona                  5m
planetscale              5m
postgresql               5m
prometheus               5m
redis-enterprise         5m
storageos                5m
strimzi-kafka-operator   5m
vault                    5m
```

#### Step 4 - Deploy the Component Operator using a Subscription

Create a `CatalogSourceConfig` and `Subscription` to the `Component Operator` in order to install the `operator` within the `operators` namespace
```bash
oc create -f deploy/olm-catalog/component-catalog-source-config.yaml -n marketplace
oc create -f deploy/olm-catalog/component-subscription.yaml -n operators
```

**Warning**: Doc of the Operator marketplace is not clear about the purpose to create a `CatalogSourceConfig` where the name is `name: installed-upstream-community-operators`

Verify if the Component Operator is up and running 
```bash
oc logs -n operators pod/component-operator-59cf6cf54-xk8mx
2019/04/04 16:26:58 Go Version: go1.11.6
2019/04/04 16:26:58 Go OS/Arch: linux/amd64
2019/04/04 16:26:58 component-operator version: unset
2019/04/04 16:26:58 component-operator git commit: b695ee1
2019/04/04 16:26:58 Registering Components
2019/04/04 16:26:58 Start the manager
```

### How to install okd 3.11 using the OCP resources

**Remark**: The OCP yaml resources are used to install olm, marketplace on ocp under `openshift-marketplace` and `openshift-operators` namespaces. They correspond or should correspond to what is deployed
on OCP4 - AWS.

Create an `oc cluster up` 3.11  and add `cluster-admin` role to the `admin` user

Install olm using the resources of the olm project created under the folder deploy/okd/manifests/latest

```bash
oc apply -f deploy/olm-catalog/olm/0000_50_olm_00-namespace.yaml
oc apply -f deploy/olm-catalog/olm/0000_50_olm_01-olm-operator.serviceaccount.yaml
oc apply -f deploy/olm-catalog/olm/0000_50_olm_02-clusterserviceversion.crd.yaml
oc apply -f deploy/olm-catalog/olm/0000_50_olm_03-installplan.crd.yaml
oc apply -f deploy/olm-catalog/olm/0000_50_olm_04-subscription.crd.yaml
oc apply -f deploy/olm-catalog/olm/0000_50_olm_05-catalogsource.crd.yaml
oc apply -f deploy/olm-catalog/olm/0000_50_olm_06-olm-operator.deployment.yaml
oc apply -f deploy/olm-catalog/olm/0000_50_olm_07-catalog-operator.deployment.yaml
oc apply -f deploy/olm-catalog/olm/0000_50_olm_08-aggregated.clusterrole.yaml
oc apply -f deploy/olm-catalog/olm/0000_50_olm_09-operatorgroup.crd.yaml
oc apply -f deploy/olm-catalog/olm/0000_50_olm_10-olm-operators.configmap.yaml
oc apply -f deploy/olm-catalog/olm/0000_50_olm_11-olm-operators.catalogsource.yaml
oc apply -f deploy/olm-catalog/olm/0000_50_olm_12-operatorgroup-default.yaml
oc apply -f deploy/olm-catalog/olm/0000_50_olm_13-packageserver.subscription.yaml
```

![olm](img/install-olm.png)

Install Openshift Marketplace using the resources of the `Marketplace` project located under the folder manifests without the resources `11_clusteroperator.yaml`, `image-references` 
and where the `node_selector` has been removed from the `10_operator.yaml` file 

```bash
oc create -f deploy/olm-catalog/marketplace/01_namespace.yaml
oc create -f deploy/olm-catalog/marketplace/02_catalogsourceconfig.crd.yaml
oc create -f deploy/olm-catalog/marketplace/03_operatorsource.crd.yaml
oc create -f deploy/olm-catalog/marketplace/04_service_account.yaml
oc create -f deploy/olm-catalog/marketplace/05_role.yaml
oc create -f deploy/olm-catalog/marketplace/06_role_binding.yaml
oc create -f deploy/olm-catalog/marketplace/07_redhat_operatorsource.cr.yaml
oc create -f deploy/olm-catalog/marketplace/08_certified_operatorsource.cr.yaml
oc create -f deploy/olm-catalog/marketplace/09_community_operatorsource.cr.yaml
oc create -f deploy/olm-catalog/marketplace/10_operator.yaml
```
![marketplace](img/install-marketplace.png)

Check if the operators appear under the `Operatorhub` screen

![marketplace](img/operatorhub.png)

Install the component operator bundle
```bash
oc create -f deploy/olm-catalog/component-operator-source.yaml -n openshift-marketplace
```

Create a subscription for the component operator
```bash
oc create -f deploy/olm-catalog/component-subscription-ocp4.yaml -n openshift-operators
```

Check if the `OperatorSources` are well installed like the `packagemanifests`
```bash
oc get opsrc --all-namespaces 
NAMESPACE               NAME                  TYPE          ENDPOINT              REGISTRY              DISPLAYNAME           PUBLISHER   STATUS      MESSAGE                                       AGE
openshift-marketplace   certified-operators   appregistry   https://quay.io/cnr   certified-operators   Certified Operators   Red Hat     Succeeded   The object has been successfully reconciled   5m
openshift-marketplace   community-operators   appregistry   https://quay.io/cnr   community-operators   Community Operators   Red Hat     Succeeded   The object has been successfully reconciled   5m
openshift-marketplace   component-operator    appregistry   https://quay.io/cnr   ch007m                Component Operator    Snowdrop    Succeeded   The object has been successfully reconciled   1m
openshift-marketplace   redhat-operators      appregistry   https://quay.io/cnr   redhat-operators      Red Hat Operators     Red Hat     Succeeded   The object has been successfully reconciled   5m
```

and

```bash
oc get packagemanifests --all-namespaces
NAMESPACE   NAME                     AGE
            component                40m
            packageserver            1h
            amq-streams              57m
            couchbase-enterprise     57m
            mongodb-enterprise       57m
            automationbroker         57m
            camel-k                  57m
            cluster-logging          57m
            cockroachdb              57m
            descheduler              57m
            elasticsearch-operator   57m
            etcd                     57m
            federation               57m
            jaeger                   57m
            kiecloud-operator        57m
            metering                 57m
            microcks                 57m
            node-network-operator    57m
            node-problem-detector    57m
            oneagent                 57m
            percona                  57m
            planetscale              57m
            postgresql               57m
            prometheus               57m
            strimzi-kafka-operator   57m
            templateservicebroker    57m
```

### Cleanup

To be reviewed !!

```bash
oc delete -f https://raw.githubusercontent.com/operator-framework/operator-marketplace/master/deploy/upstream/02_catalogsourceconfig.crd.yaml
oc delete -f https://raw.githubusercontent.com/operator-framework/operator-marketplace/master/deploy/upstream/03_operatorsource.crd.yaml
oc delete -f https://raw.githubusercontent.com/operator-framework/operator-marketplace/master/deploy/upstream/04_service_account.yaml
oc delete -f https://raw.githubusercontent.com/operator-framework/operator-marketplace/master/deploy/upstream/05_role.yaml
oc delete -f https://raw.githubusercontent.com/operator-framework/operator-marketplace/master/deploy/upstream/06_role_binding.yaml
oc delete -f https://raw.githubusercontent.com/operator-framework/operator-marketplace/master/deploy/upstream/07_operator.yaml

oc delete -f https://raw.githubusercontent.com/operator-framework/operator-marketplace/master/deploy/examples/upstream.operatorsource.cr.yaml

oc delete -f https://raw.githubusercontent.com/operator-framework/operator-marketplace/master/deploy/upstream/01_namespace.yaml

oc delete -f https://raw.githubusercontent.com/operator-framework/operator-lifecycle-manager/master/deploy/upstream/quickstart/olm.yaml 
```
