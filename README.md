# Developer Experience Runtime Operator

[![CircleCI](https://circleci.com/gh/snowdrop/component-operator/tree/master.svg?style=shield)](https://circleci.com/gh/snowdrop/component-operator/tree/master)

Table of Contents
=================
   * [Introduction](#introduction)
   * [Prerequisites](#prerequisites)
   * [Setup](#setup)
      * [Local cluster using Minikube](#local-cluster-using-minikube)
   * [Installation of the Operator](#installation-of-the-operator)
   * [How to play with it](#how-to-play-with-it)
   * [A more complex scenario](#a-more-complex-scenario)
   * [Switch from Development to Build/Prod mode](#switch-from-development-to-buildprod-mode)
   * [A cool demo](#a-cool-demo)
   * [Cleanup](#cleanup)

## Introduction

Modern applications designed using the `Microservices` pattern or the [12-factor](https://12factor.net/) methodology requires when they will be deployed
on a Kubernetes cluster a strong Developer Experience able to deal with the different and sometimes complex Kubernetes resources needed.

This project has been developed in order to help them and to `simplify` the process to deploy such applications.

This is the reason why, to enhance the Developer Experience on Kubernetes, we have designed different [Custom Resources - CR](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) and
a Kubernetes [Operator](https://enterprisersproject.com/article/2019/2/kubernetes-operators-plain-english) able to perform the following tasks:
- Install different `runtimes` (aka microservices) such as `Spring Boot, Vert.x, Thorntail, Quarkus or Nodejs`
- Manage the relations which exist between the `Microservices` using a CR `link` to consume by example a REST endpoint or
- To setup a CR `capability` on the platform to access a backend like a database: postgresql

The `Custom Resource` contains `METADATA` information about the framework/language to be used to either:
- Configure the strategy that we want to adopt to deploy the application: `Development mode` or `Building/Prod mode`
- Select the container image to be used to launch the application: java for `Spring Boot, Eclipse Vert.x, Thorntail`; node for `nodejs` 
- Configure the `Microservice` in order to inject `env var, secret, ...`
- Create a service or capability such as database: postgresql

## Prerequisites

In order to use the DevExp Runtime Operator and the CRs, it is needed to install [Tekton Pipelines](https://tekton.dev/) and [KubeDB](http://kubedb.com) Operators.
We assume that you have installed a K8s cluster as of starting from Kubernetes version 1.12.

### Local cluster using Minikube

Install using `brew tool` on `MacOS` the following software
```bash
brew cask install minikube
brew install kubernetes-cli
brew install kubernetes-helm
```

Next, create a `K8s` cluster where `ingress` and `dashboard` addons are enabled
```bash
minikube config set vm-driver virtualbox
minikube config set cpus 4
minikube config set kubernetes-version v1.14.0
minikube config set memory 6000
minikube addons enable ingress
minikube addons enable dashboard
minikube start
```

When `minikube` has started, initialize `Helm` to install on the cluster `Tiller`

```bash
helm init
until kubectl get pods -n kube-system -l name=tiller | grep 1/1; do sleep 1; done
kubectl create clusterrolebinding tiller-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:default
```

Install Tekton Pipelines technology
```bash
kubectl apply -f https://storage.googleapis.com/tekton-releases/previous/v0.4.0/release.yaml
```

Install the `KubeDB` operator and its `PostgreSQL` catalog supporting different database versions
```bash
KUBEDB_VERSION=0.12.0
helm repo add appscode https://charts.appscode.com/stable/
helm repo update
helm install appscode/kubedb --name kubedb-operator --version ${KUBEDB_VERSION} \
  --namespace kubedb --set apiserver.enableValidatingWebhook=false,apiserver.enableMutatingWebhook=false
```

Wait till the Operator has started before to install the Catalog
```bash
TIMER=0
until kubectl get crd elasticsearchversions.catalog.kubedb.com memcachedversions.catalog.kubedb.com mongodbversions.catalog.kubedb.com mysqlversions.catalog.kubedb.com postgresversions.catalog.kubedb.com redisversions.catalog.kubedb.com || [[ ${TIMER} -eq 60 ]]; do
  sleep 5
  TIMER=$((TIMER + 1))
done
```

Install the PostgreSQL catalog.
```bash
helm install appscode/kubedb-catalog --name kubedb-catalog --version ${KUBEDB_VERSION} \
  --namespace kubedb --set catalog.postgres=true,catalog.elasticsearch=false,catalog.etcd=false,catalog.memcached=false,catalog.mongo=false,catalog.mysql=false,catalog.redis=false
```

## Installation of the DevExp Runtime Operator

Deploy the `Cluster Role`, `Role Binding`, `CRDs`, `ServiceAccount` and `Operator` within the namespace `component-operator` 

```bash
kubectl create ns operators
kubectl apply -n operators -f deploy/sa.yaml
kubectl apply -f deploy/cluster-role.yaml
kubectl apply -f deploy/user-rbac.yaml
kubectl apply -f deploy/cluster-role-binding.yaml
kubectl apply -f deploy/crds/capability_v1alpha2.yaml
kubectl apply -f deploy/crds/component_v1alpha2.yaml
kubectl apply -f deploy/crds/link_v1alpha2.yaml
kubectl apply -n operators -f deploy/operator.yaml
```

Wait till the Operator's pod is ready and running before to continue
```bash
until kubectl get pods -n operators -l name=component-operator | grep 1/1; do sleep 1; done
```

Control if the operator is runnign correctly
```bash
pod_id=$(kubectl get pods -n operators -l name=component-operator -o=name)
kubectl logs $pod_id -n operators
```

Enjoy now to play with the DevExp Runtime Operator !

## How to play with it

The process is pretty simple and trivial and is about creating a custom resource, one by microservice or runtime to be deployed.
So, create first a `demo` namespace
```bash
kubectl create ns demo
```

and next, create a `component's yaml` file with the following information
```bash
echo "
apiVersion: devexp.runtime.redhat.com/v1alpha2
kind: Component
metadata:
  name: spring-boot
spec:
  runtime: spring-boot
  deploymentMode: dev" | kubectl apply -n demo -f -
```

Verify if the component has been well created by executing the following kubectl command
```bash
kc get components -n demo
NAME          RUNTIME       VERSION   AGE   MODE   STATUS   MESSAGE   REVISION
spring-boot   spring-boot             11s   dev                       
```
When, the DevExp operator will read the content of the Custom Resource `Component`, then it will create several K8s resources that you can discover if you execute the following command
```bash
kubectl get all,pvc,component
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

To cleanup the project installed (component)
```bash  
$ kubectl delete component --all -n demo 
``` 
  
## A more complex scenario   

In order to play with a more complex scenario where we would like to install 2 components: `frontend`, `backend` and a database's service from the Ansible Broker's catalog
like also the `links` needed to update the `DeploymentConfig`, then you should execute the following commands at the root of the github project within a terminal

```bash
kubectl apply -f examples/demo/component-client.yml
kubectl apply -f examples/demo/component-link-env.yml
kubectl apply -f examples/demo/component-crud.yml
kubectl apply -f examples/demo/component-service.yml
kubectl apply -f examples/demo/component-link.yml
```  
  
## Switch from Development to Build/Prod mode

The existing operator supports to switch from the `inner` or development mode (where code must be pushed to the development's pod) to the `outer` mode (responsible to perform a `s2i` build 
deployment using a SCM project). In this case, a container image will be created from the project compiled and next a new deploymentConfig will be created in order to launch the 
runtime.

In order to switch between the 2 modes, execute the following operations: 

Decorate the `Component CRD yaml` file with the following values in order to specify the git info needed to perform a Build, like the name of the component to be selected to switch from
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
  
Patch the component when it has been deployed to switch from the `inner` to the `outer` deployment mode
  
```bash
kubectl patch cp fruit-backend-sb -p '{"spec":{"deploymentMode":"outerloop"}}' --type=merge
```   

## A cool demo
  
## TODO: to be reviewed 

The deployment of an application will consist in to create a `Component` yaml resource file defined according to the 
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
  

### Cleanup the Operator resources

```bash
kubectl delete -n operators -f deploy/sa.yaml
kubectl delete -f deploy/cluster-role.yaml
kubectl delete -f deploy/user-rbac.yaml
kubectl delete -f deploy/cluster-role-binding.yaml
kubectl delete -f deploy/crds/capability_v1alpha2.yaml
kubectl delete -f deploy/crds/component_v1alpha2.yaml
kubectl delete -f deploy/crds/link_v1alpha2.yaml
kubectl delete -n operators -f deploy/operator.yaml
```