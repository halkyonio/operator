# Component CRD & Kubernetes operator

[![CircleCI](https://circleci.com/gh/snowdrop/component-operator/tree/master.svg?style=shield)](https://circleci.com/gh/snowdrop/component-operator/tree/master)

Table of Contents
=================
  * [Introduction](#introduction)
  * [For the users](#for-the-users)
     * [How to play with the Component operator locally](#how-to-play-with-the-component-operator-locally)
     * [A more complex scenario](#a-more-complex-scenario)
     * [Switch from Development to Build/Prod mode](#switch-from-development-to-buildprod-mode)
     * [How to install the operator on a cluster](#how-to-install-the-operator-on-a-cluster)
     * [Cleanup](#cleanup)

   
## Introduction

The purpose of this project is to develop a Kubernetes `Custom Resource Definition` called `Component CRD` and [Kubernetes Operator](https://goo.gl/D8iE2K) able to install a Microservice Application such as `Spring Boot` on a cloud platform: `Kubernetes` or `OpenShift`
or to deploy a Service using the help of Service Broker Catalog.

The CRD contains `METADATA` information about the framework/language to be used to either:
- Configure the strategy that we want to adopt to code and install the application: `Development mode` or `Building/Prod mode`
- Select the container image to be used to launch the application: java for `Spring Boot, Eclipse Vert.x, Thorntail`; node for `nodejs` 
- Configure the microservice in order to inject `env var, secret, ...`
- Create a service selected from a Service Broker Catalog

The deployment or installation of an application in a namespace will consist in to create a `Component` yaml resource file defined according to the 
[Component API spec](https://github.com/snowdrop/component-operator/blob/master/pkg/apis/component/v1alpha2/component_types.go#L11).

```bash
apiVersion: devexp.runtime.redhat.com/v1alpha2
kind: Component
metadata:
  name: spring-boot-demo
spec:
  # Strategy used by the operator to install the Kubernetes resources as DevMode = dev or BuildMode = outerloop
  deploymentMode: dev
  # Runtime type that the operator will map with a docker image (java, node, ...)
  runtime: spring-boot
  version: 1.5.16
  # To been able to create a Kubernetes Ingress resource OR OpenShift Route
  exposeService: true
  envs:
    - name: SPRING_PROFILES_ACTIVE
      value: openshift-catalog
```

When this `Custom resource` will be processed by the Kubernetes API Server and published, then the `Component operator` will be notified and will execute different operations to create:
- For the `runtime` a development's pod running a `supervisor's daemon` able to start/stop the application [**[1]**](https://github.com/snowdrop/component-operator/blob/master/pkg/pipeline/dev/install.go#L56) and where we can push a `uber jar` file compiled locally, 
- A Service using the OpenShift Automation Broker and the Kubernetes Service Catalog [**[2]**](https://github.com/snowdrop/component-operator/blob/master/pkg/pipeline/servicecatalog/install.go),
- `EnvVar` section for the development's pod [**[3]**](https://github.com/snowdrop/component-operator/blob/master/pkg/pipeline/link/link.go#L56).

**Remark**: The `Component` Operator can be deployed on `Kubernetes >= 1.11` or `OpenShift >= 3.11`.

## For the users

### TODO - To be reviewed and updated

- Login to the cluster using a user having admin cluster role
```bash
oc login "https://<cluster_ip>:8443" -u admin -p admin
```

- Install Tekton Pipelines technology
```bash
kubectl apply -f https://storage.googleapis.com/tekton-releases/previous/v0.4.0/release.yaml
```

- Disable TLS verification
```bash
kubectl config set-cluster <cluster_ip>:8443 --insecure-skip-tls-verify=false
```

- Install KubeDB operator and Postgresql catalog
```bash
kubectl create ns kubedb
curl -fsSL https://raw.githubusercontent.com/kubedb/cli/0.12.0/hack/deploy/kubedb.sh \
    | bash -s -- --namespace=kubedb --install-catalog=postgres --enable-validating-webhook=false --enable-mutating-webhook=false

OR

KUBEDB_VERSION=0.12.0
helm init
until kubectl get pods -n kube-system -l name=tiller | grep 1/1; do sleep 1; done
kubectl create clusterrolebinding tiller-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:default

helm repo add appscode https://charts.appscode.com/stable/
helm repo update
helm install appscode/kubedb --name kubedb-operator --version ${KUBEDB_VERSION} \
--namespace kubedb --set apiserver.enableValidatingWebhook=false,apiserver.enableMutatingWebhook=false

TIMER=0
until kubectl get crd elasticsearchversions.catalog.kubedb.com memcachedversions.catalog.kubedb.com mongodbversions.catalog.kubedb.com mysqlversions.catalog.kubedb.com postgresversions.catalog.kubedb.com redisversions.catalog.kubedb.com || [[ ${TIMER} -eq 60 ]]; do
  sleep 2
  TIMER=$((TIMER + 1))
done

helm install appscode/kubedb-catalog --name kubedb-catalog --version ${KUBEDB_VERSION} \
--namespace kubedb --set catalog.postgres=true,catalog.elasticsearch=false,catalog.etcd=false,catalog.memcached=false,catalog.mongo=false,catalog.mysql=false,catalog.redis=false
```

- Verify operator installation

To check if KubeDB operator pods have started, run the following command:
```
kubectl get pods --all-namespaces -l app=kubedb --watch
```

Once the operator pods are running, you can cancel the above command by typing Ctrl+C.
Now, to confirm CRD groups have been registered by the operator, run the following command:

```
kubectl get crd -l app=kubedb
```
Now, you are ready to create your first database using KubeDB

- Deploy the resources within the namespace `component-operator` 

```bash
kubectl create ns component-operator
kubectl apply -f https://raw.githubusercontent.com/snowdrop/component-operator/master/deploy/sa.yaml
kubectl apply -f https://raw.githubusercontent.com/snowdrop/component-operator/master/deploy/cluster-rbac.yaml
kubectl apply -f https://raw.githubusercontent.com/snowdrop/component-operator/master/deploy/user-rbac.yaml
kubectl apply -f https://raw.githubusercontent.com/snowdrop/component-operator/master/deploy/crds/capability_v1alpha2.yaml
kubectl apply -f https://raw.githubusercontent.com/snowdrop/component-operator/master/deploy/crds/component_v1alpha2.yaml
kubectl apply -f https://raw.githubusercontent.com/snowdrop/component-operator/master/deploy/crds/link_v1alpha2.yaml
kubectl apply -f https://raw.githubusercontent.com/snowdrop/component-operator/master/deploy/operator.yaml
```

- Give the `priveleged` security context to the serviceaccount `postgres-db` used by the KubeDB operator
```bash
oc project test
oc adm policy add-scc-to-user privileged -z postgres-db
```

- Give also the security context `privileged` to the serviceaccount `build-bot` used to run the Task of the pod. Git ve it to this serviceaccount the role to `edit`
```bash
oc adm policy add-scc-to-user privileged -z build-bot
oc adm policy add-role-to-user edit -z build-bot
```

## Clean up

```
curl -fsSL https://raw.githubusercontent.com/kubedb/cli/0.12.0/hack/deploy/kubedb.sh \
    | bash -s -- --uninstall -n kubedb
```

error: unable to retrieve the complete list of server APIs: mutators.kubedb.com/v1alpha1: the server is currently unable to handle the request, validators.kubedb.com/v1alpha1: the server is currently unable to handle the request


### How to play with the Component operator locally

- Log on to an OpenShift cluster >=3.10 with cluster-admin rights
- Create a namespace `component-operator`
  ```bash
  $ oc new-project component-operator
  ```

- Deploy the resources: service account, rbac and crd definition
  ```bash
  $ oc create -f deploy/sa.yaml
  $ oc create -f deploy/rbac.yaml
  $ oc create -f deploy/crds/component_v1alpha2.yaml
  ```

- Start the Operator locally using the `Main` go file
  ```bash
  $ oc new-project my-spring-app
  $ OPERATOR_NAME=component-operator WATCH_NAMESPACE=my-spring-app KUBERNETES_CONFIG=$HOME/.kube/config go run cmd/manager/main.go
  
- In a separate terminal create a component's yaml file with the following information
  ```bash
  echo "
  apiVersion: devexp.runtime.redhat.com/v1alpha2
  kind: Component
  metadata:
    name: my-spring-boot
  spec:
    runtime: spring-boot
    deploymentMode: dev" | oc apply -f -
  ```

- Check if the `operator` has created the following kubernetes resources, part of the `dev` deployment mode
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
  component.devexp.runtime.redhat.com/my-spring-boot   spring-boot                                               46s
  
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
  
### Switch from Development to Build/Prod mode

The existing operator supports to switch from the `inner` or development mode (where code must be pushed to the development's pod) to the `outer` mode (responsible to perform a `s2i` build 
deployment using a SCM project). In this case, a container image will be created from the project compiled and next a new deploymentConfig will be created in order to launch the 
runtime.

In order to switch between the 2 modes, execute the following operations: 

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
  
  **Remark**: When the maven project does not contain multi modules, then replace the name of the folder / module with `.` using the annotation `app.openshift.io/git-dir`
  
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
  oc delete -f deploy/crds/component-v1alpha2.yaml
  oc delete -f deploy/operator.yaml
  oc delete -f deploy/rbac.yaml
  oc delete -f deploy/sa.yaml
  ```