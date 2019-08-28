# Halkyon Operator

[![CircleCI](https://circleci.com/gh/halkyonio/operator/tree/master.svg?style=shield)](https://circleci.com/gh/halkyonio/operator/tree/master)
[![GitHub release](https://img.shields.io/github/v/release/halkyonio/operator.svg)](https://github.com/halkyonio/operator/releases/latest)
[![Licensed under Apache License version 2.0](https://img.shields.io/github/license/halkyonio/operator?maxAge=2592000)](https://www.apache.org/licenses/LICENSE-2.0)

Table of Contents
=================
  * [Introduction](#introduction)
  * [Prerequisites](#prerequisites)
     * [Local cluster using Minikube](#local-cluster-using-minikube)
  * [Installation of the Halkyon Operator](#installation-of-the-halkyon-operator)
  * [How to play with it](#how-to-play-with-it)
  * [A Real demo](#a-real-demo)
  * [Cleanup the Operator resources](#cleanup-the-operator-resources)
  * [Support](#support)

## Introduction

Deploying modern micro-services applications that utilize the [12-factor](https://12factor.net/) guidelines to Kubernetes is difficult, mainly due to the host of different and complex Kubernetes Resources involved. In such scenarios developer experience becomes very important. 

This projects aims to tackle said complexity and vastly **simplify** the process of deploying micro-service applications to Kubernetes.

By providing several, easy-to-use Kubernetes [Custom Resources - CR](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) and
an [Operator](https://enterprisersproject.com/article/2019/2/kubernetes-operators-plain-english) to handle them, the Halkyon project provides the following features:
- Install micro-services (`components` in Halkyon's parlance) utilizing `runtimes` such as Spring Boot, Vert.x, Thorntail, Quarkus or Nodejs, serving as base building blocks for your application
- Manage the relations between the different components of the using a `link` CR allowing one micro-service for example to consume a REST endpoint provided by another
- Deploy various infrastructure services like a database which are bound to a `component` via the `capability` CR.

Custom resources contain metadata about the framework/language to be used to either:
- Configure the deployment strategy used to deploy the application: `Development mode` or `Building/Prod mode`
- Select the container image to be used to launch the application: java for `Spring Boot, Eclipse Vert.x, Thorntail`; node for `nodejs` 
- Configure the `component` in order to inject `env var, secret, ...`
- Create a service or capability such as a database

As an example of its usage, by deploying a CR such as the following to a cluster that has the Halkyon operator installed, Halkyon is able to add a micro-service to your project, to generate a Route URL providing access to the service outside the cluster and to add environment variables to the created pod.
 
```bash
apiVersion: halkyon.io/v1beta1
kind: Component
metadata:
  name: spring-boot-demo
spec:
  # Strategy used by the operator to install the Kubernetes resources using :
  # 1) the Dev mode where we can push the uber jar file of the microservices
  # compiled locally or 
  # 2) Build Mode where we want to delegate to the operator the responsibility
  # to build the code from the git repository and to push to a local
  # docker registry the image
  deploymentMode: dev
  # Runtime type that the operator will map with a docker image (java, nodejs, ...)
  runtime: spring-boot
  version: 2.1.16
  # To been able to create a Kubernetes Ingress resource or OpenShift Route
  exposeService: true
  envs:
  - name: SPRING_PROFILES_ACTIVE
    value: openshift-catalog
```

**Remarks**: 
- The Halkyon Operator runs top of `Kubernetes >= 1.13` or `OpenShift >= 3.11`.
- You can find more information about the Custom Resources and their fields under the project https://github.com/halkyonio/api like also the others `CRs` supported : `link` and `capability`

## Prerequisites

In order to use the Halkyon Operator and the CRs, it is needed to install [Tekton Pipelines](https://tekton.dev/) and [KubeDB](http://kubedb.com) Operators.
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

## How to play with it

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

## A Real demo
  
To play with a `real example` and discover the different features currently supported, we have created within the directory `demo` a project containing 
2 micro-services: a Spring Boot REST client calling a Service exposed by a Spring Boot backend application which access a postgresql database.

So jump [here](demo/README.md) in order to see in action How we enhance the Developer Experience on Kubernetes ;-)

## Cleanup the Operator resources

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

## Support

If you need support, reach out to us via [zulip].

If you run into issues, don't hesitate to raise an [issue].

Follow us on [twitter].

[website]: https://halkyion.io
[zulip]: https://snowdrop.zulipchat.com/#narrow/stream/207165-halkyon
[issue]: https://github.com/halkyonio/operator/issues/new
[twitter]: https://twitter.com/halkyonio
