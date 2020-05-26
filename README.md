# Halkyon Operator: get back to the halcyon days of local development in a modern kubernetes setting!

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
   * [Pre-requisites](#pre-requisites)
      * [Local cluster using Minikube](#local-cluster-using-minikube)
   * [Installing the Halkyon Operator](#installing-the-halkyon-operator)
      * [How to play with it](#how-to-play-with-it)
      * [A Real demo](#a-real-demo)
      * [Cleanup the Operator resources](#cleanup-the-operator-resources)
   * [Compatibility matrix](#compatibility-matrix)
   * [Support](#support)

## Introduction

Deploying modern micro-services applications that comply with the [12-factor](https://12factor.net/) guidelines to Kubernetes is difficult, mainly due to the host of different and complex Kubernetes Resources involved. In such scenarios developer experience becomes very important. 

This projects aims to tackle said complexity and vastly **simplify** the process of deploying micro-service applications to Kubernetes and get back to the halcyon days of local development! :sunglasses:

By providing several, easy-to-use Kubernetes [Custom Resources (CRs)](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) and
an [Operator](https://enterprisersproject.com/article/2019/2/kubernetes-operators-plain-english) to handle them, the Halkyon project provides the following features:
- Install micro-services (`components` in Halkyon's parlance) utilizing `runtimes` such as Spring Boot, Vert.x, Thorntail, Quarkus or Nodejs, serving as base building blocks for your application
- Deploy various infrastructure services, like databases, that components can then use to implement their functionality, via the `capability` CR
- Record the dependencies components need to operate using a contract-based approach where a component describes the set of capabilities the component requires and/or provides to other components

The Halkyon Operator requires `Kubernetes >= 1.13` or `OpenShift >= 3.11`.

## Key concepts

We will explain Halkyon's key concepts using the example of a simple, modern application: a `frontend` application connecting
to a `backend` application via a REST endpoint. This application, in turns, uses the services of a PostgreSQL database.

Such an application, though simple, will require several Kubernetes resources in order to be deployed on a Kubernetes cluster. 
Furthermore, several development iterations are usually required to make the application production ready.

In Halkyon parlance, both `frontend` and `backend` micro-services are `components` of our application. The PostgreSQL database
is a `capability` used by the `backend` `component`. Components define which `capabilities` they require to function and which
they provide to other `components` via contract-like declarations. Halkyon can then match, automatically or on-demand, `components`
and `capabilities` based on the cluster state.

For example, the `backend` component declares requiring a PostgreSQL database capability, expliciting its requirements but also 
declares providing a REST endpoint capability that can be used by other components, which is exactly what the `frontend` needs.
Halkyon takes care of exposing the `backend` cluster URL and injecting that information into the `frontend` component when 
the binding is requested.

Similarly, should a PostgreSQL database `capability` matching the `backend` requirements be deployed on the cluster, Halkyon could
bind it the `backend` `component` thus providing the `component` with the database connection information without the user having
to figure out how to do so. 

Halkyon is extensible and it is not too hard to implement new capabilities that can be dynamically added to the operator, without
requiring rebuilding the core operator. More details about Halkyon's 
[extension mechanism](https://github.com/halkyonio/operator-framework/blob/master/README.adoc#plugin-architecture-overview) 
can be found in the [Halkyon framework project](https://github.com/halkyonio/operator-framework) 
[documentation](https://github.com/halkyonio/operator-framework/). 

Information about `components` and `capabilities` are materialized by custom resources in Halkyon. We can create the
manifests for these custom resources which, once processed by the remote cluster, will be handled by the Halkyon operator to 
create the appropriate Kubernetes/OpenShift resources for you, so you can focus on your application architecture as opposed to 
wondering how it might translate to Kubernetes `pods` or `deployments`.

**Remark**: you can view the full description of the CRs and their API in the [associated Halkyon API project](https://github.com/halkyonio/api).

### Component

A component represents a micro-service, i.e. part of an application to be deployed. The `Component` custom resource provides a
simpler to fathom abstraction over what's actually required at the Kubernetes level to deploy and optionally expose the 
micro-service outside of the cluster. In fact, when a `component` is deployed to a Halkyon-enabled cluster, the Halkyon operator
will create these resources:
- `Deployment`,
- `Service`,
- `PersistentVolumeClaim`,
- `Ingress` or `Route` on OpenShift if the component is exposed.

#### Runtime

You can already see how Halkyon reduces the cognitive load on developers since there is no need to worry about the low-level 
details by focusing on the salient aspects of your component: what runtime does it need to run, does it need to be exposed outside
of the cluster and on what port. Theses aspects are captured along with less important ones in the custom resource fields: 
`runtime` (and `version`), `exposeService` and `port`. The `runtime` name will condition which container image will be used to 
run the application. 

Of note, the java-based runtimes currently use a specific image which allows us to do builds from source as well
as run binaries. For more information about this image, please take a look at 
https://github.com/halkyonio/container-images/blob/master/README.md#hal-maven-jdk8-image. Runtimes are currently defined using
custom resources and it's therefore easy to deploy new runtimes that Halkyon can use. We are, however, considering switching to 
using [devfiles](https://devfile.github.io/website/devfile/).

#### Mode

Halkyon offers two deployment modes, controlled by the `deploymentMode` field of the custom resource: `dev` 
(for "development") and `build`, `dev` being the default mode if none is specified explicitly.

The `dev` mode sets the environment in such a way that the pod where your application is 
deployed doesn't need to be restarted when the code changes. On the contrary, the pod contains an init container exposing a 
server that can listen to commands so that your application executable can be restarted or re-compiled after updates without 
needing to restart the whole pod or generate a new container image which allows for faster turn-around. 

The `build` mode uses the Tekton Pipeline Operator in order to build of a new image for your application. How the image is built
is controlled by the `buildConfig` field of the `component` custom resource where you need to minimally specify the url of the 
git repository to be used as basis for the code (`url` field). You can also specify the precise git reference to use (`ref` field)
or where to find the actual code to build within the repository using the `contextPath` and `moduleDirName` fields.

#### Provided and required capabilities

As described earlier, `components` specify the set of `capabilities` they require to function as well as the set of `capabilities`
they provide for other `components` to leverage. A `component` CR defines its contract in the `capabilities` section of its `spec`, 
which contains, as expected, two arrays: `requires` and `provides`.

For example, here is how the `backend` component described above would declare its contract:

```yaml
capabilities:
    provides:
    - name: backend-endpoint
      spec:
        category: api
        parameters:
        - name: context
          value: /api/fruits
        type: rest-component
        version: "1"
    requires:
    - autoBindable: true
      name: db
      spec:
        category: database
        type: Postgres
        version: "10.6"
```  

The above defines one required and one provided `capabilities`. Halkyon will match capabilities based on their category, type 
and, optionally, version. This declaration means that the `backend` component requires a `10.6` PostgreSQL database capability,
which is marked as `autoBindable`, meaning that Halkyon will bind to the first capability matching the requirements found on the
cluster. This is useful in a development environment to go faster but is probably not a good idea on a production cluster! :)
If a user knows which `capability` to bind against, they can explicitly request it using the `boundTo` field of the required 
capability definition. Halkyon will then attempt to bind to the specified capability if available. Of note, this field will be
automatically set by Halkyon when an automatic binding occurs so that the matching process is subsequently bypassed.

The provided capability defines that the `backend` component provides an API REST endpoint on the `/api/fruits` context as 
specified by the `context` parameter.

#### Reference

For more details on the fields of the Component custom resource, please refer to 
[its API](https://github.com/halkyonio/api/blob/master/component/v1beta1/types.go).

**Examples**:

`DeploymentMode: dev`
```yaml
apiVersion: halkyon.io/v1beta1
kind: Component
metadata:
  name: backend
spec:
  buildConfig:
    ref: ""
    url: ""
  capabilities:
    provides:
    - name: backend-endpoint
      spec:
        category: api
        parameters:
        - name: context
          value: /api/fruits
        type: rest-component
        version: "1"
    requires:
    - autoBindable: true
      name: db
      spec:
        category: database
        type: Postgres
        version: "10.6"
  deploymentMode: dev
  envs:
  - name: SPRING_PROFILES_ACTIVE
    value: kubernetes
  exposeService: true
  port: 8080
  runtime: spring-boot
  version: 2.1.13.RELEASE
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

### Capability 

A capability corresponds to a service that the micro-service will consume on the platform. The Halkyon operator then uses this 
information to configure the service. Capabilities are identified by the combination of their `category` which represents the 
general class of configurable services, further identified by a more specific `type` (which could be construed as a sub-category)
and a version for the `category/type` combination. The service is then configured using a list of name/value `parameters`.

Capabilities are implemented as plugins and are therefore independent of Halkyon's core, meaning that any user can extend 
Halkyon by developing new capabilities but also that Halkyon's core is not burdened by the dependencies a given capability might
bring, thus making things easier to manage. See the documentation on Halkyon's 
[extension mechanism](https://github.com/halkyonio/operator-framework/blob/master/README.adoc#plugin-architecture-overview) for 
more details. 

For example, Halkyon uses the `[KubeDB](https://kubedb.com)` operator to handle the database category. The plugin implementation
can be found in the [`kubedb-capability`](https://github.com/halkyonio/kubedb-capability) project.

For more details on the fields of the Capability custom resource, please refer to 
[its API](https://github.com/halkyonio/api/blob/master/capability/v1beta1/types.go).

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

## Pre-requisites

In order to use the Halkyon Operator and the CRs, the [Tekton Pipelines](https://tekton.dev/) operator needs to be installed on the cluster.
Capabilities might have additional requirements. For example, the [KubeDB](http://kubedb.com) operator is required for the 
`kubedb-capability` plugin. We assume that you have installed a cluster with Kubernetes version equals to 1.13 or newer.

### Local cluster using `minikube`

Install using Homebrew on `macOS` the following software:
```bash
brew cask install minikube
brew install kubernetes-cli
brew install kubernetes-helm
```

Next, create a Kubernetes cluster where `ingress` and `dashboard` addons are enabled
```bash
minikube config set cpus 4
minikube config set kubernetes-version v1.14.0
minikube config set memory 8000
minikube addons enable ingress
minikube addons enable dashboard
minikube addons enable registry
minikube start
```

Install Tekton Pipelines:
```bash
kubectl apply -f https://storage.googleapis.com/tekton-releases/pipeline/previous/v0.9.1/release.yaml
```

Install the `KubeDB` operator and the catalog of the databases using the following bash script as described within the `kubedb` [doc](https://kubedb.com/docs/0.12.0/setup/install/):
```bash
kubectl create ns kubedb
curl -fsSL https://raw.githubusercontent.com/kubedb/cli/0.12.0/hack/deploy/kubedb.sh \
    | bash -s -- --namespace=kubedb
```

**Note**: To remove it, use the following parameters `kubedb.sh --namespace=kubedb --uninstall --purge`

## Installing the Halkyon Operator

Install the Halkyon operator within the `operators` namespace:
```bash
./scripts/halkyon.sh operators install yes
```

Wait until the Operator's pod is ready and running before continuing:
```bash
until kubectl get pods -n operators -l name=halkyon-operator | grep 1/1; do sleep 1; done
```

Control if the operator is running correctly:
```bash
pod_id=$(kubectl get pods -n operators -l name=halkyon-operator -o=name)
kubectl logs $pod_id -n operators
```

You can also use the operator bundle promoted on [operatorhub.io](https://operatorhub.io/operator/halkyon).

### Running a new version of the Halkyon operator on an already-setup cluster

Let's assume that you've already installed Halkyon on a cluster (i.e. kubedb and tekton operators are setup and the Halkyon 
resources are deployed on the cluster) but that you want to build a new version of the operator to, for example, test a bug fix.

The easiest way to do so is to scale down the deployment associated with the operator down to 0 replicas in your cluster. Of 
course, you need to make sure that no one else is relying on that operator running on the cluster! Assuming the above 
installation, you can do so by:

```bash
kubectl scale --replicas=0 -n operators $(kubectl get deployment -n operators -o name)
```

You can then compile and run the Halkyon operator locally. This assumes you have set up a Go programming environment, know 
your way around using Go:

```bash
go get halkyon.io/operator
cd $GOPATH/src/halkyon.io/operator
make
```

You will then want to run the operator locally so that the cluster can call it back when changes are detected to Halkyon 
resources. You will do so by running the operator watching the specific namespace where you want to test changes. Watching a 
specific namespace ensures that your locally running instance of the operator doesn't impact users in different namespaces (and
also insures that you don't see changes made to resources in other namespaces that you might not be interested in):

```bash
WATCH_NAMESPACE=<the name of your namespace here>; go run ./cmd/manager/main.go
```

Enjoy the Halkyon Operator!

### How to play with it

Deploy the operator as defined within the [Operator Doc](https://github.com/halkyonio/operator#installing-the-halkyon-operator)

First create a `demo` namespace:
```bash
kubectl create ns demo
```
Next, create a `component` yml file with the following information within your maven java project:
```bash
apiVersion: halkyon.io/v1beta1
kind: Component
metadata:
  name: spring-boot
spec:
  runtime: spring-boot
  version: 2.1.6.RELEASE
  deploymentMode: dev
  port: 8080
```

Deploy it: 
```bash
kubectl apply -n demo -f my-component.yml
```

Verify if the component has been deployed properly:
```bash
kubectl get components -n demo
NAME          RUNTIME       VERSION        AGE   MODE   STATUS    MESSAGE                                                            REVISION
spring-boot   spring-boot   2.1.6.RELEASE  14s   dev    Pending   pod is not ready for component 'spring-boot' in namespace 'demo'                        
```

**Remark** Don't worry about the initial status as downloading the needed images from an external docker registry could take time!

```bash
kubectl get components -n demo   
NAME          RUNTIME       VERSION         AGE   MODE   STATUS   MESSAGE   REVISION
spring-boot   spring-boot   2.1.6.RELEASE   36m   dev    Ready    Ready              
```

The Halkyon operator will then use the content of the `component` custom resource to create the Kubernetes resources needed to 
materialize your application on the cluster. You can see all these resources by executing the following command:
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

Package your Java Application `mvn package` and push the `uber` java file.
```bash
kubectl cp target/my-component-1.0-SNAPSHOT.jar POD_NAME:/deployments/app.jar -n demo
```

**Remark**: You can get the pod name or pod id using this command : `kubectl get pods -l component_cr=spring-boot -o name` where
you pass as `component_cr` label, the component name. Remove the `pod/` prefix from the name. E.g: `pod/spring-boot-747995b4db-hqxhd` -> `spring-boot-747995b4db-hqxhd`

Start your application within the pod
```bash
kubectl exec POD_NAME -n demo /var/lib/supervisord/bin/supervisord ctl start run
```

**Important**: We invite you to use our [`Hal` companion tool](https://github.com/halkyonio/hal#2-deploy-the-component) as it will create and push the code source or binary without having to worry about the kubectl command syntax ;-)

Enrich your application with additional `Component`, `Link` them or deploy a `Capability` database using the supported CRs for your different microservices.
To simplify your life even more when developing Java applications, add [Dekorate]( https://dekorate.io) to your project to automatically generate the YAML resources for your favorite runtime !

You can now cleanup the project:
```bash  
kubectl delete component --all -n demo 
```

### A Real demo

To play with a more real-world example and discover the different features currently supported, we have implemented the application 
we took as an example in the [Key Concepts section](#key-concepts). You can find it in the [`demo` directory](./demo).

So jump [here](demo/README.md) to see in action how Halkyon enhances the Developer Experience on Kubernetes :wink:

### Cleanup the operator resources

To remove the operator from your favorite Kubernetes cluster, then execute the following command:
```bash
./scripts/halkyon.sh operators delete
```

## Compatibility matrix

|                     | Kubernetes >= 1.13 | OpenShift 3.x | OpenShift 4.x | KubeDB 0.12 | Tekton v0.9.x | 
|---------------------|--------------------|---------------|---------------|-------------|---------------|
| halkyon v0.1.x      | ✓                  | ✓             | ✓             | ✓           | ✓             |

## Support

If you need support, reach out to us via [zulip](https://snowdrop.zulipchat.com/#narrow/stream/207165-halkyon).

If you run into issues or if you have questions, don't hesitate to raise an [issue](https://github.com/halkyonio/operator/issues/new).

Follow us on [twitter](https://twitter.com/halkyonio).