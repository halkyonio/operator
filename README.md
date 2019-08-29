# Halkyon Operator

[![CircleCI](https://circleci.com/gh/halkyonio/operator/tree/master.svg?style=shield)](https://circleci.com/gh/halkyonio/operator/tree/master)
[![GitHub release](https://img.shields.io/github/v/release/halkyonio/operator.svg)](https://github.com/halkyonio/operator/releases/latest)
[![Licensed under Apache License version 2.0](https://img.shields.io/github/license/halkyonio/operator?maxAge=2592000)](https://www.apache.org/licenses/LICENSE-2.0)

Table of Contents
=================
   * [Introduction](#introduction)
   * [Key concepts](#key-concepts)
      * [Component](#component)
      * [Link](#link)
      * [Capability](#capability)
   * [Prerequisites](#prerequisites)
      * [Local cluster using Minikube](#local-cluster-using-minikube)
   * [Installation of the Halkyon Operator](#installation-of-the-halkyon-operator)
      * [How to play with it](#how-to-play-with-it)
      * [A Real demo](#a-real-demo)
      * [Cleanup the Operator resources](#cleanup-the-operator-resources)
   * [Compatibility matrix](#compatibility-matrix)
   * [Support](#support)

## Introduction

Deploying modern micro-services applications that comply with the [12-factor](https://12factor.net/) guidelines to Kubernetes is difficult, mainly due to the host of different and complex Kubernetes Resources involved. In such scenarios developer experience becomes very important. 

This projects aims to tackle said complexity and vastly **simplify** the process of deploying micro-service applications to Kubernetes.

By providing several, easy-to-use Kubernetes [Custom Resources - CR](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) and
an [Operator](https://enterprisersproject.com/article/2019/2/kubernetes-operators-plain-english) to handle them, the Halkyon project provides the following features:
- Install micro-services (`components` in Halkyon's parlance) utilizing `runtimes` such as Spring Boot, Vert.x, Thorntail, Quarkus or Nodejs, serving as base building blocks for your application
- Manage the relations between the different components using `link` CR allowing one micro-service for example to consume a REST endpoint provided by another
- Deploy various infrastructure services like a database which are bound to a `component` via the `capability` CR.

The Halkyon Operator requires `Kubernetes >= 1.13` or `OpenShift >= 3.11`.

## Key concepts

A modern application, which is defined as a collection of micro-services as depicted hereafter

![Composition](component-operator-demo.png)

will require several Kubernetes resources in order to be deployed on a kubernetes cluster. Furthermore several development iterations are usually required to make the 
application production ready.

When, during the analysis phase, the developers will discuss the final picture about the solution to be designed, they
will certainly adopt this very straightforward convention (aka Fluent DSL) to express the different micro-services, their relations and ultimately
the services needed.

`(from:componentA).(to:componentB).(to:serviceA)`

The `ComponentA` and `ComponentB` correspond respectively to by example a Spring Boot application `fruit-client-sb` and `fruit-backend-sb`.

The relation `from -> to` indicates that we will `reference` the `ComponentA`  with the `ComponentB` using a `Link`.

The `link`'s purpose is to inject as `Env var(s)` the information required to by example configure the `HTTP client` of the `ComponentA` to access the 
`ComponentB` which exposes a `HTTP endpoint` that the client `ComponentA`  could use to access the `CRUD` operations.

To make `ComponentB` capable of communicating with a database, a `link` will also be setup in order to pass using the database `Secret` the parameters which are needed to configure a Java Datasource.

This is, for that reason, that such `entities/concepts` are also proposed by Halkyon using different Kubernetes Custom Resources: `Component`, `Link` and  `Capabiltity`.

**Remark**: you can view the full description of the CR and its API under the project `https://github.com/halkyonio/api`.

### Component

**Definition**:

The component represents a micro-service of an application to be deployed and contains the following information defined within `spec` section: 
- `deploymentMode`: to configure the deployment strategy used : `Development` or `Building/Prod`.
- `runtime`: to select the container image to be used to launch the application. For the `Spring Boot, Eclipse Vert.x, Thorntail`, an `OpenJDK8` image will be used while for a `Node` runtime
   that will be a `nodejs` image.
- `version`: the version of the runtime as defined within the local maven, gradle, ... project   
- `exposeService`: to create an ingress resource (on kubernetes) or a route resource (on openshift) to access the endpoint of the service outside of the cluster   
- `8080`: runtime port to be used to access the endpoint's service
- `envs`: to specify `env vars` needed by the runtime.

By deploying a `Component`, the Halkyon operator will create these Kubernetes resources: 
- `Deployment`,
- `Service`,
- `PVC`,
- `Route` or `Ingress` (optional)

The deployment strategy is used to manage 2 distinct deployments on the cluster:
- `development`
- `build`

The `development` mode will be used to create a pod including as init container, a supervisord application which exposes different `commands` that a tool or a command executed
within the pod can trigger in order to execute `assemble`, `run`, `compile`. These commands are mapped to the s2i executables packaged within the Java OpenJDK S2i image or S2i nodejs images.

A typical development scenario will consist of first creating a project, designing a micro-service and when the code is ready to be tested on the cluster an artifact will be created (such as an `uber-jar`). 
Next the artifact needs to be pushed to the pod and finally to call the command launching the application. 

The `build` mode will under the hood use the Tekton Pipeline Operator in order to create a Tekton Pod of executing, depending on to the `BuildConfig` type, the build of the image
using the steps defined within the `Task`. Currently, we only support a `S2I` build to generate the runtime image.
To configure the build, the `Component` must include additional parameters defined within the `BuildConfig` field where:
- `type`: refers to the Pipeline or build strategy to be done. The default value is `s2i`
- `url`: is the url of the git repo to be cloned. Example : `https://github.com/halkyonio/operator.git`
- `ref`: is the branch or git tag to be cloned. Default value is `master`
- `contextPath`: allows to specify within the project cloned the path containing the project of the application. Default is `.`
- `moduleDirName`: is the directory name of the maven module to compile. Default value is `.`

**Example**

`DeploymentMode: dev`  
  
```yaml
apiVersion: halkyon.io/v1beta1
kind: Component
metadata:
  name: spring-boot-demo
spec:
  deploymentMode: dev
  runtime: spring-boot
  version: 2.1.16
  exposeService: true
  port: 8080
  envs:
  - name: SPRING_PROFILES_ACTIVE
    value: openshift-catalog
```

`DeploymentMode: build`
```yaml
apiVersion: "halkyon.io/v1beta1"
kind: "Component"
metadata:
  labels:
    app: "fruit-backend-sb"
  name: "fruit-backend-sb"
spec:
  deploymentMode: "build"
  runtime: "spring-boot"
  version: "2.1.6.RELEASE"
  exposeService: true
  buildConfig:
    type: "s2i"
    url: "https://github.com/halkyonio/operator.git"
    ref: "master"
    contextPath: "demo/"
    moduleDirName: "fruit-backend-sb"
  port: 8080
```
  
### Link

**Definition**:

The link represents the information which is needed by a micro-service to access a HTTP endpoint exposed by another microservice or a service like a database.
It supports 2 different types: `Secret` or `Env` which are used by the Operator in order to inject using the `componentName` referenced
the information provided.

A Link contains the following parameters:
- `componentName`: is target `component` where the information should be injected. This component corresponds to a kubernetes `Deployment` resource
- `type`: The `Secret` type allows the operator to search about a kubernetes Secret according to the `ref` and next to inject into the `Deployment`, withint
the `.spec.container` part the values within the `EnvFrom` field. The `Env` type will be used to enrich the  `.spec.container` part of the `Deployment` with the 
`Envs` vars defined within the array `envs
- `ref`: Kubernetes secret reference to look for within the namespace
- `envs`: list of env variables defined as `name` and `value` pairs  

**Example**:

`Secret`
```yaml
apiVersion: "halkyon.io/v1beta1"
kind: "Link"
metadata:
  name: "link-to-database"
spec:
  componentName: "fruit-backend-sb"
  type: "Secret"
  ref: "postgres-db-config"
```

`Envs`
```yaml
apiVersion: "halkyon.io/v1beta1"
kind: "Link"
metadata:
  name: "link-to-fruit-backend"
spec:
  componentName: "fruit-client-sb"
  type: "Env"
  ref: ""
  envs:
  - name: "ENDPOINT_BACKEND"
    value: "http://fruit-backend-sb:8080/api/fruits"
```

### Capability 

**Definition**

A capability corresponds to a service that the micro-service will consume on the platform. It will be used as input by the operator to configure the service
as defined hereafter:
- `category`: it represents a capability supported by the operator. We only support for the moment the `database` category
- `type`: According to the `category`, this field represents by example the type of the database to be deployed. `PostgreSQL` is only supported for the moment. 
- `version`: identify the version of the `type` to be installed.
- `parameters`: list of `name` and `value` pairs used to create the secret of the service

The Halkyon operator uses for the `category` database, the `[KubeDB](https://kubedb.com)` operator to delegate the creation/installation of the pod of the database within the namespace of the 
user.

**Example**:

`PostgreSQL Database`
```yaml
apiVersion: "halkyon.io/v1beta1"
  kind: "Capability"
  metadata:
    name: "postgres-db"
  spec:
    category: "database"
    type: "postgres"
    version: "10"
    parameters:
    - name: "DB_USER"
      value: "admin"
    - name: "DB_PASSWORD"
      value: "admin"
    - name: "DB_NAME"
      value: "sample-db"
```

## Prerequisites

In order to use the Halkyon Operator and the CRs, the [Tekton Pipelines](https://tekton.dev/) and [KubeDB](http://kubedb.com) Operators need to be installed on the cluster.
We assume that you have installed a K8s cluster as of starting from Kubernetes version 1.13.

### Local cluster using Minikube

Install using Homebrew on `macOS` the following software:
```bash
brew cask install minikube
brew install kubernetes-cli
brew install kubernetes-helm
```

Next, create a Kubernetes cluster where `ingress` and `dashboard` addons are enabled
```bash
minikube config set vm-driver virtualbox
minikube config set cpus 4
minikube config set kubernetes-version v1.14.0
minikube config set memory 6000
minikube addons enable ingress
minikube addons enable dashboard
minikube addons enable registry
minikube start
```

When `minikube` has started, initialize `Helm` to install on the cluster `Tiller`

```bash
helm init
until kubectl get pods -n kube-system -l name=tiller | grep 1/1; do sleep 1; done
kubectl create clusterrolebinding tiller-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:default
```

Next, edit the `/etc/hosts` file of minikube vm in order to specify the IP address of the docker registry daemon
as the docker client, when it will fetch from the docker registry images pushed, will use name `kube-registry.kube-system` to call the 
docker server

```bash
kc get svc/registry -n kube-system -o jsonpath={.spec.clusterIP}
<IP_ADDRESS> 
minikube ssh
echo '<IP_ADDRESS> kube-registry.kube-system kube-registry.kube-system.svc kube-registry.kube-system.svc.cluster.local' | sudo tee -a /etc/hosts
```

Install Tekton Pipelines technology
```bash
kubectl apply -f https://storage.googleapis.com/tekton-releases/previous/v0.5.2/release.yaml
```

Install the `KubeDB` operator and its `PostgreSQL` catalog supporting different database versions
```bash
KUBEDB_VERSION=0.12.0
helm repo add appscode https://charts.appscode.com/stable/
helm repo update
helm install appscode/kubedb --name kubedb-operator --version ${KUBEDB_VERSION} \
  --namespace kubedb --set apiserver.enableValidatingWebhook=true,apiserver.enableMutatingWebhook=true
```

Wait until the Operator has started before to install the Catalog
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

## Installation of the Halkyon Operator

Deploy the `Cluster Role`, `Role Binding`, `CRDs`, `ServiceAccount` and `Operator` within the namespace `operators` 

```bash
kubectl create ns operators
kubectl apply -n operators -f deploy/sa.yaml
kubectl apply -f deploy/cluster-role.yaml
kubectl apply -f deploy/user-rbac.yaml
kubectl apply -f deploy/cluster-role-binding.yaml
kubectl apply -f deploy/crds/capability.yaml
kubectl apply -f deploy/crds/component.yaml
kubectl apply -f deploy/crds/link.yaml
kubectl apply -n operators -f deploy/operator.yaml
```

Wait till the Operator's pod is ready and running before to continue
```bash
until kubectl get pods -n operators -l name=halkyon-operator | grep 1/1; do sleep 1; done
```

Control if the operator is running correctly
```bash
pod_id=$(kubectl get pods -n operators -l name=halkyon-operator -o=name)
kubectl logs $pod_id -n operators
```

Enjoy the Halkyon Operator!

### How to play with it

The process is pretty simple and is about creating a custom resource, one by micro-service or runtime to be deployed.
So create first a `demo` namespace
```bash
kubectl create ns demo
```
and next create using your favorite editor, a `component` resource with the following information:
```bash
apiVersion: halkyon.io/v1beta1
kind: Component
metadata:
  name: spring-boot
spec:
  runtime: spring-boot
  deploymentMode: dev
```

Deploy it 
```bash
kubectl apply -n demo -f my-component.yaml
```

Verify if the component has been well created by executing the following `kubectl` command
```bash
kubectl get components -n demo
NAME          RUNTIME       VERSION   AGE   MODE   STATUS    MESSAGE                                                            REVISION
spring-boot   spring-boot             14s   dev    Pending   pod is not ready for component 'spring-boot' in namespace 'demo'                        
```

**Remark** Don't worry about the status which is reported the first time as downloading the needed images from a nexternal docker registry could take time !

```bash
kubectl get components -n demo   
NAME          RUNTIME       VERSION   AGE     MODE   STATUS   MESSAGE   REVISION
spring-boot   spring-boot             2m19s   dev    Ready              
```

When, the Halkyon operator will read the content of the Custom Resource `Component`, then it will create several K8s resources that you can discover if you execute the following command
```bash
kubectl get pods,services,deployments,pvc -n demo
NAME                              READY   STATUS    RESTARTS   AGE
pod/spring-boot-6d9475f4c-c9w2z   1/1     Running   0          4m18s

NAME                  TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
service/spring-boot   ClusterIP   10.104.75.68   <none>        8080/TCP   4m18s

NAME                                READY   UP-TO-DATE   AVAILABLE   AGE
deployment.extensions/spring-boot   1/1     1            1           4m18s

NAME                                        STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/m2-data-spring-boot   Bound    pvc-dab00dfe-a2f6-11e9-98d1-08002798bb5f   1Gi        RWO            standard       4m18s
```

You can now cleanup the project as we will not deploy additional micro-services. So you can cleanup the project installed (component)
```bash  
kubectl delete component --all -n demo 
```

### A Real demo
  
To play with a `real example` and discover the different features currently supported, we have created within the directory `demo` a project containing 
2 micro-services: a Spring Boot REST client calling a Service exposed by a Spring Boot backend application which access a postgresql database.

So jump [here](demo/README.md) in order to see in action How we enhance the Developer Experience on Kubernetes ;-)

### Cleanup the Operator resources

To clean the operator deployed on your favorite Kubernetes cluster, then execute the following kubectl commands:

```bash
kubectl delete -n operators -f deploy/sa.yaml
kubectl delete -f deploy/cluster-role.yaml
kubectl delete -f deploy/user-rbac.yaml
kubectl delete -f deploy/cluster-role-binding.yaml
kubectl delete -f deploy/crds/capability.yaml
kubectl delete -f deploy/crds/component.yaml
kubectl delete -f deploy/crds/link.yaml
kubectl delete -n operators -f deploy/operator.yaml
```

## Compatibility matrix

|                     | Kubernetes 1.13 | OpenShift 3.x | OpenShift 4.x | KubeDB 0.12 | Tekton v0.5.x| 
|---------------------|-----------------|---------------|---------------|-------------|--------------|
| halkyon v0.1.x      | ✓               | ✓             | ✓             | ✓           | ✓            |

## Support

If you need support, reach out to us via [zulip].

If you run into issues, don't hesitate to raise an [issue].

Follow us on [twitter].

[website]: https://halkyion.io
[zulip]: https://snowdrop.zulipchat.com/#narrow/stream/207165-halkyon
[issue]: https://github.com/halkyonio/operator/issues/new
[twitter]: https://twitter.com/halkyonio
