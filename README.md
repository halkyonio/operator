# Developer Experience Runtime Operator

[![CircleCI](https://circleci.com/gh/snowdrop/component-operator/tree/master.svg?style=shield)](https://circleci.com/gh/snowdrop/component-operator/tree/master)

Table of Contents
=================
  * [Introduction](#introduction)
  * [Prerequisites](#prerequisites)
     * [Local cluster using Minikube](#local-cluster-using-minikube)
  * [Installation of the DevExp Runtime Operator](#installation-of-the-devexp-runtime-operator)
  * [How to play with it](#how-to-play-with-it)
  * [A Real demo](#a-real-demo)
  * [Cleanup the Operator resources](#cleanup-the-operator-resources)

## Introduction

Modern applications designed using the `Microservices` pattern or the [12-factor](https://12factor.net/) methodology requires when they will be deployed
on a Kubernetes cluster a strong Developer Experience able to deal with the different and sometimes complex Kubernetes resources needed.

This project has been developed in order to help you and to `simplify` the process to deploy such applications.

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

To define what a microservices is, if a URL should be created to access it from outside of the cluster or simply to specify some `ENV` vars to be used by the 
container's image when the pod is created, then create a `component.yaml` file containing the following information: 
 
```bash
apiVersion: devexp.runtime.redhat.com/v1alpha2
kind: Component
metadata:
  name: spring-boot-demo
spec:
  # Strategy used by the operator to install the Kubernetes resources using a Dev mode where we can push the uber jar file of the microservices compiled locally
  # or Build Mode where we want to delegate to the operator the responsibility to build the code from the git repository and to push to a local kubernetes registry the image
  deploymentMode: dev
  # Runtime type that the operator will map with a docker image (java, nodejs, ...)
  runtime: spring-boot
  version: 1.5.16
  # To been able to create a Kubernetes Ingress resource or OpenShift Route
  exposeService: true
  envs:
  - name: SPRING_PROFILES_ACTIVE
    value: openshift-catalog
```

**Remarks**: 
- The `DevExp Runtime` Operator runs top of `Kubernetes >= 1.11` or `OpenShift >= 3.11`.
- You can find more information about the `Custom Resources` and their `fields` under this folder `pkg/apis/component/v1alpha2` like also the others `CRs` supported : `link` and `capability`

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

The process is pretty simple and is about creating a custom resource, one by microservice or runtime to be deployed.
So create first a `demo` namespace
```bash
kubectl create ns demo
```
and next create using your favorite editor a `component's yaml` file with the following information
```bash
apiVersion: devexp.runtime.redhat.com/v1alpha2
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

Verify if the component has been well created by executing the following kubectl command
```bash
kubectl get components -n demo
NAME          RUNTIME       VERSION   AGE   MODE   STATUS    MESSAGE                                                            REVISION
spring-boot   spring-boot             14s   dev    Pending   pod is not ready for component 'spring-boot' in namespace 'demo'                        
```

**Remark** Don't worry about the status which is reported the first time as downloading the first time the needed images from external docker registry could take time !

```bash
kubectl get components -n demo   
NAME          RUNTIME       VERSION   AGE     MODE   STATUS   MESSAGE   REVISION
spring-boot   spring-boot             2m19s   dev    Ready              
```

When, the DevExp operator will read the content of the Custom Resource `Component`, then it will create several K8s resources that you can discover if you execute the following command
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

You can now cleanup the project as we will not deploy a Java Microservices or delegate the build to the Operator. So cleanup the project installed (component)
```bash  
kubectl delete component --all -n demo 
```

## A Real demo
  
To play with a `real example` and discover the different features currently supported, we have created within the directory `demo` a project containing 
2 microservices: a Spring Boot REST client calling a Service exposed by a Spring Boot backend application which access a postgresql database.

So jump [here](demo/README.md) in order to see in action How we enhance the Developer Experience on Kubernetes ;-)

## Cleanup the Operator resources

To clean the operator deployed on your favorite Kubernetes cluster, then execute the following kubectl commands:

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