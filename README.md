## Component's kubernetes operator

[![CircleCI](https://circleci.com/gh/snowdrop/component-operator/tree/master.svg?style=shield)](https://circleci.com/gh/snowdrop/component-operator/tree/master)

Table of Contents
=================

  * [For the users](#for-the-users)
    * [How to play with the Component operator locally](#how-to-play-with-the-component-operator-locally)
    * [A more complex scenario](#a-more-complex-scenario)
    * [Switch fromm inner to outer mode](#switch-fromm-inner-to-outer)
    * [How to install the operator on an cluster](#how-to-install-the-operator-on-the-cluster)
    * [Cleanup the operator](#cleanup)
  * [For the developers only](#for-the-developers-only)
      * [How To create the operator, crd](#how-to-create-the-operator-crd)
      * [How to install the Operator on OCP4](#how-to-install-the-operator-on-ocp4)
         * [Package and install the Operator on Quay.io as Application](#package-and-install-the-operator-on-quayio-as-application)
         * [Deploy on OCP4](#deploy-on-ocp4)

## For the users

### How to play with the Component operator locally

- Log on to an OpenShift cluster >=3.10 with cluster-admin rights
- Create a namespace `component-operator`
  ```bash
  $ oc new-project component-operator
  ```

- Deploy the resources : service account, rbac and crd definition
  ```bash
  $ oc create -f deploy/sa.yaml
  $ oc create -f deploy/rbac.yaml
  $ oc create -f deploy/crds/crd.yaml
  ```

- Start the Operator locally using the `Main` go file
  ```bash
  $ oc new-project my-spring-app
  $ OPERATOR_NAME=component-operator WATCH_NAMESPACE=my-spring-app KUBERNETES_CONFIG=$HOME/.kube/config go run cmd/manager/main.go
  
- In a separate terminal create a component's yaml file with the following information
  ```bash
  echo "
  apiVersion: component.k8s.io/v1alpha1
  kind: Component
  metadata:
    name: my-spring-boot
  spec:
    runtime: spring-boot
    deploymentMode: innerloop" | oc apply -f -
  ```

- Check if the `operator` has created the following kubernetes resources, part of the `innerloop` deployment mode
  ```bash
  oc get all,pvc,component
  NAME                         READY     STATUS    RESTARTS   AGE
  pod/my-spring-boot-1-nrszv   1/1       Running   0          41s
  
  NAME                                     DESIRED   CURRENT   READY     AGE
  replicationcontroller/my-spring-boot-1   1         1         1         44s
  
  NAME                     TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)    AGE
  service/my-spring-boot   ClusterIP   172.30.73.5   <none>        8080/TCP   45s
  
  NAME                                                REVISION   DESIRED   CURRENT   TRIGGERED BY
  deploymentconfig.apps.openshift.io/my-spring-boot   1          1         1         image(copy-supervisord:latest),image(dev-runtime-spring-boot:latest)
  
  NAME                                                     DOCKER REPO                                                     TAGS      UPDATED
  imagestream.image.openshift.io/copy-supervisord          docker-registry.default.svc:5000/demo/copy-supervisord          latest    45 seconds ago
  imagestream.image.openshift.io/dev-runtime-spring-boot   docker-registry.default.svc:5000/demo/dev-runtime-spring-boot   latest    45 seconds ago
  
  NAME                                        RUNTIME       VERSION   SERVICE   TYPE      CONSUMED BY   AGE
  component.component.k8s.io/my-spring-boot   spring-boot                                               46s
  
  NAME                                           STATUS    VOLUME    CAPACITY   ACCESS MODES   STORAGECLASS   AGE
  persistentvolumeclaim/m2-data-my-spring-boot   Bound     pv005     1Gi        RWO                           46s
  ```

- To cleanup the project installed (component)
  ```bash  
  $ oc delete components,route,svc,is,pvc,dc --all=true && 
  ``` 
  
### A more complex scenario   

In order to play with a more complex scenario where we would like to install 2 components: `frontend`, `backend` and a database's service from the Ansible Broker's catalog
like also the `links` needed to update the `DeploymentConfig`, then you should execute the following commands at the root of the github project within a terminal

  ```bash
  oc apply -f examples/demo/component-client.yml
  oc apply -f examples/demo/component-link-env.yml
  oc apply -f examples/demo/component-crud.yml
  oc apply -f examples/demo/component-service.yml
  oc apply -f examples/demo/component-link.yml
  ```  
  
### Switch from inner to outer

The existing operator supports to switch from the `inner` or development mode (where code must be pushed to the development's pod) to the `outer` mode (responsible to perform a `s2i` build 
deployment using a SCM project). In this case, a container image will be created from the project compiled and next a new deploymentConfig will be created in order to launch the 
runtime.

In order to switch, execute the following operations: 

- Decorate the `Component CRD yaml` file with the following values in order to specify the git info needed to perform a Build, like the name of the component to be selected to switch from
  the dev loop to the outer loop

  ```bash
   annotations:
     app.openshift.io/git-uri: https://github.com/snowdrop/component-operator-demo.git
     app.openshift.io/git-ref: master
     app.openshift.io/git-dir: fruit-backend-sb
     app.openshift.io/artifact-copy-args: "*.jar"
     app.openshift.io/runtime-image: "fruit-backend-sb"
     app.openshift.io/component-name: "fruit-backend-sb"
     app.openshift.io/java-app-jar: "fruit-backend-sb-0.0.1-SNAPSHOT.jar"
  ``` 
  
  **Remark** : When the maven project does not contain multi modules, then replace the name of the folder / module with `.` using the annotation `app.openshift.io/git-dir`
  
- Patch the component when it has been deployed to switch from the `inner` to the `outer` deployment mode
  
  ```bash
  oc patch cp fruit-backend-sb -p '{"spec":{"deploymentMode":"outerloop"}}' --type=merge
  ```   
  
### How to install the operator on a cluster

In the previous section, the operator was launched locally using a `go SDK`. If you would like to install it as a `container` on an OpenShift cluster, then it is required to build 
the container image using the `operator-sdk` tool [1], to push the image and next to install the operator using a `Deployment` kubernetes resource as defined hereafter

  ```bash
  operator-sdk build quay.io/snowdrop/component-operator
  docker push quay.io/snowdrop/component-operator
  oc create -f deploy/operator.yaml
  ``` 
  
[1] https://github.com/operator-framework/operator-sdk   

### Cleanup

  ```bash
  oc delete -f deploy/cr.yaml
  oc delete -f deploy/crds/crd.yaml
  oc delete -f deploy/operator.yaml
  oc delete -f deploy/rbac.yaml
  oc delete -f deploy/sa.yaml
  ```   
  
## For the developers only
  
### How To create the operator, crd

Instructions followed to create the Component's CRD, operator using the `operator-sdk`'s kit

- Execute this command within the `$GOPATH/github.com/$ORG/` folder is a terminal
  ```bash
  operator-sdk new component-operator --api-version=component.k8s.io/v1alpha1 --kind=Component --skip-git-init
  operator-sdk add api --api-version=component.k8s.io/v1alpha1 --kind=Component 
  ```
  using the following parameters 

  Name of the folder to be created : `component-operator`
  Api Group Name   : `component.k8s.io`
  Api Version      : `v1alpha1`
  Kind of Resource : `Component` 

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
  oc create -f deploy/crds/crd.yaml
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
  oc delete -f deploy/cr.yaml
  oc delete -f deploy/crds/crd.yaml
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

### How to install the Operator on OCP4

#### Package and install the Operator on Quay.io as Application

Install the [tool](https://github.com/operator-framework/operator-courier) `operator-courier`

    pip3 install operator-courier
    
Verify your operator's bundle using the tool

    export BUNDLE_DIR="deploy/olm-catalog/bundle"
    operator-courier verify $BUNDLE_DIR
    WARNING:operatorcourier.validate:csv spec.icon not defined    

Next, get from `quay.io` an Authentication token using your quay's username OR robot username/pwd to access your namespace
Edit the file `deploy/olm-catalog/quay_login.json` and add your username/pwd.
Next, execute the following curl request to get a `basic Y2gwMDdtK...A="` response

    export AUTH_TOKEN=$(curl -X POST -H "Content-Type: application/json" -d @deploy/olm-catalog/quay_login.json https://quay.io/cnr/api/v1/users/login | jq -r '.token')
    
Push finally the application

    export QUAY_ORG="quay_organization (e.g ch007m)"
    export REPOSITORY="component-operator"
    export RELEASE="0.1.0"
    operator-courier push $BUNDLE_DIR $QUAY_ORG $REPOSITORY $RELEASE "$AUTH_TOKEN"
    
#### Deploy on OCP4

Log on to an ocp4 cluster as a cluster-admin role user
Deploy the `OperatorSource` in order to install from `Quay.io/app` the bundle of the operator

    oc apply -f deploy/olm-catalog/operator.source.yaml -n openshift-marketplace  

Next, subscribe to the olm operator 

TODO: Copmponent-operator pod is created under `openshift-marketplace` but the log of the operator does not seem to correspond to our operator !!

### How to install the Operator using an OLM

TODO: To be reviewed as the installation was failing !!

#### Step 1 - Install the new OpenShift console

- Launch an ocp/okd 3.11 cluster locally
- Log on as `cluster-admin` role user
- Git clone the new console and build it
```
git clone https://github.com/talamer/console.git && cd console
./build.sh
```

- Launch the proxy console locally

```
source ./contrib/oc-environment.sh
./bin/bridge
```

- Install the Operator Lifecycle Manager

- As `cluster-admin` role user, install the Operator Lifecycle Manager (OLM) using this command:
```
oc create -f https://raw.githubusercontent.com/operator-framework/operator-lifecycle-manager/master/deploy/upstream/quickstart/olm.yaml 
```

#### Step 2 - Clone/build the Operator

- Check that you have a docker daemon running locally and export the DOCKER_HOST
```
git clone git@github.com:redhat-developer/devopsconsole-operator.git $GOPATH/src/github.com/redhat-developer/devopsconsole-operator
cd $GOPATH/src/github.com/redhat-developer/devopsconsole-operator
make build 
```

#### Step 3 - Push the operator image to quay

- Login to the docker quay hub using your Quay user / pwd
```
docker login -u="USER" -p="PWD" quay.io
```

- Build the docker image and export it
```
export QUAY_USER="ch007m"
operator-sdk build quay.io/$QUAY_USER/devopsconsole-operator
docker push quay.io/$QUAY_USER/devopsconsole-operator
```
**WARNING** Remember to set the operator's repository visibility to public at quay.io

#### Step 4 - Generate the OLM catalog files
```
operator-sdk olm-catalog gen-csv --csv-version 0.1.0
```

**WARNING**. Rename or copy the file `deploy/cluster_role.yaml to `role.yaml`

#### Step 5 - Clone/build the Operator Registry
```
git clone git@github.com:sbose78/community-operators.git
cd community-operators
cp -pR rhd-operators ../devopsconsole-operator
cd ../devopsconsole-operator

wget https://raw.githubusercontent.com/sbose78/community-operators/master/rhd.Dockerfile
cp Dockerfile.builder Dockerfile.builder_OPERATOR
cp rhd.Dockerfile Dockerfile.builder
operator-sdk build quay.io/$QUAY_USER/operator-registry
cp Dockerfile.builder_OPERATOR Dockerfile.builder

docker push quay.io/$QUAY_USER/operator-registry
Remember to set the operator registry visibility to public at quay.io
```

#### Step 6 - Create Catalog Source Resource

Navigate to Administration->CRDs->CatalogSources and create a new CatalogSource by importing YAML:

Insert this text:
```
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  creationTimestamp: '2019-03-18T13:34:06Z'
  generation: 1
  name: rhd-operatorhub-catalog
  namespace: olm
  resourceVersion: '152627'
  selfLink: >-
    /apis/operators.coreos.com/v1alpha1/namespaces/olm/catalogsources/rhd-operatorhub-catalog
  uid: 80950cfc-4982-11e9-a02b-5254003ac934
spec:
  displayName: Community Operators
  image: 'quay.io/{username}/operator-registry:new'
  publisher: RHD Operator Hub
  sourceType: grpc
status:
  lastSync: '2019-03-18T13:43:31Z'
  registryService:
    createdAt: '2019-03-18T13:43:31Z'
    port: '50051'
    protocol: grpc
    serviceName: rhd-operatorhub-catalog
    serviceNamespace: olm
```

The operator should now be present in the list of available operators:
And the console registry pod should be running:
And the DevConsole operator should be visible

Then, create a subscription. Navigate to Operator Management->Operator Subscriptions->Create Subscription

Insert this text as YAML:
```
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  generateName: devopsconsole-
  namespace: operators
spec:
  source: rhd-operatorhub-catalog
  sourceNamespace: olm
  name: devopsconsole
  startingCSV: devopsconsole.v0.1.0
  channel: alpha
```

#### Step 7 - Verify that the Operator is present

```
NAME                                                                     CREATED AT
catalogsources.operators.coreos.com                                      2019-03-15T17:56:00Z
clusterserviceversions.operators.coreos.com                              2019-03-15T17:55:59Z
gitsources.devopsconsole.openshift.io                                    2019-03-18T15:03:54Z
installplans.operators.coreos.com                                        2019-03-15T17:56:00Z
openshiftwebconsoleconfigs.webconsole.operator.openshift.io              2019-03-15T17:50:15Z
operatorgroups.operators.coreos.com                                      2019-03-15T17:56:00Z
servicecertsigneroperatorconfigs.servicecertsigner.config.openshift.io   2019-03-15T17:47:44Z
subscriptions.operators.coreos.com                                       2019-03-15T17:56:00Z

oc get clusterserviceversions
NAME                       DISPLAY                    VERSION     REPLACES   PHASE
devopsconsole.v0.1.0       OpenShift DevOps Console   0.1.0                  Succeeded
packageserver.v4.0.0-olm   Package Server             4.0.0-olm              Succeeded
```
