# Developer's section

   * [Prerequisites](#prerequisites)
   * [Install the go project](#install-the-go-project)
   * [Build the supervisor and java s2i images](#build-the-supervisor-and-java-s2i-images)
      * [Common step](#common-step)
      * [Supervisord image](#supervisord-image)
      * [Java S2I image](#java-s2i-image)
   * [Scenario to be executed to validate changes before to commit](#scenario-to-be-executed-to-validate-changes-before-to-commit)
      * [Test 0 : Build executable and test it](#test-0--build-executable-and-test-it)
      * [Test 1 : source -&gt; compile -&gt; run](#test-1--source---compile---run)
      * [Test 2 : binary -&gt; run](#test-2--binary---run)
      * [Test 3 : debug](#test-3--debug)
      * [Test 4 : source -&gt; compile -&gt; kill pod -&gt; compile again (m2 repo is back again)](#test-4--source---compile---kill-pod---compile-again-m2-repo-is-back-again)
      * [Test 5 : build (code not yet finalized as image is build bit no deployment is available)](#test-5--build-code-not-yet-finalized-as-image-is-build-bit-no-deployment-is-available)
      * [Test 6 : Scaffold a project](#test-6--scaffold-a-project)


# Prerequisites

 - Go Lang : [>=1.9](https://golang.org/doc/install)
 - [GOWORKSPACE](https://golang.org/doc/code.html#Workspaces) variable defined
 - [minishift](https://docs.okd.io/latest/minishift/)

# Install the go project

Download and install the `spring-boot-cloud-devex's github project` within your `$GOPATH`'s directory

```bash
cd $GOPATH/src
go get github.com/snowdrop/spring-boot-cloud-devex
cd spring-boot-cloud-devex
```   

# Build locally the tool

```bash
make build
sudo cp sb /usr/local/bin
```

# Build the supervisor & java s2i images

## Common step
 
Export the Docker ENV var (`DOCKER_HOST, ....`) to access the docker daemon
```bash
eval $(minishift docker-env)
```
  
## Supervisord image  

To build the `copy-supervisord` docker image containing the `go supervisord` application, then follow these instructions

**WARNING**: In order to build a multi-stages docker image, it is required to install [imagebuilder](https://github.com/openshift/imagebuilder) 
as the docker version packaged with minishift is too old and doesn't support such multi-stage option !

```bash
cd supervisord
imagebuilder -t copy-supervisord:latest .
```
  
Tag the docker image and push it to `quay.io`

```bash
TAG_ID=$(docker images -q copy-supervisord:latest)
docker tag $TAG_ID quay.io/snowdrop/supervisord
docker login quai.io
docker push quay.io/snowdrop/supervisord
```
  
## Java S2I image

Due to permissions's issue to access the folder `/tmp/src`, the Red Hat OpenJDK1.8 S2I image must be enhanced to add the permission needed for the group `0`

Here is the snippet's part of the Dockerfile

```docker
USER root
RUN mkdir -p /tmp/src/target

RUN chgrp -R 0 /tmp/src/ && \
    chmod -R g+rw /tmp/src/
```

**IMPORTANT**: See this doc's [part](https://docs.openshift.org/latest/creating_images/guidelines.html#openshift-specific-guidelines) for more info about `Support Arbitrary User IDs`

Execute such commands to build the docker image of the `Java S2I` and publish it on `Quay.io`
 
```bash
docker build -t spring-boot-http:latest .
TAG_ID=$(docker images -q <username>/spring-boot-http:latest)
docker tag $TAG_ID quay.io/snowdrop/spring-boot-s2i
docker push quay.io/snowdrop/spring-boot-s2i
```  

# Scenario to be executed to validate changes before to commit

## Test 0 : Build executable and test it

- Log on to an OpenShift cluster with an `admin` role
- Open or create the following project : `spring-boot-cloud-devexd`
- Move under the `spring-boot` folder and run these commands

```bash
oc delete --force --grace-period=0 all --all
oc delete pvc/m2-data

go build -o sb *.go
export PATH=$PATH:$(pwd)
export CURRENT=$(pwd)

cd spring-boot
mvn clean package

sb init -n cloud-demo
sb push --mode binary
sb exec start
sb exec stop
cd $CURRENT
```

## Test 1 : source -> compile -> run

- Log on to an OpenShift cluster with an `admin` role
- Open or create the following project : `cloud-demo`
- Move under the `spring-boot` folder and run these commands

```bash
oc delete --force --grace-period=0 all --all
oc delete pvc/m2-data

go build -o sb *.go
export PATH=$PATH:$(pwd)
export CURRENT=$(pwd)

cd spring-boot
sb init -n cloud-demo
sb push --mode source
sb compile
sb exec start
cd $CURRENT
```

- Execute this command within another terminal

```bash
URL="http://$(oc get routes/spring-boot-http -o jsonpath='{.spec.host}')"
curl $URL/api/greeting
```

## Test 2 : binary -> run

- Log on to an OpenShift cluster with an `admin` role
- Open or create the following project : `cloud-demo`
- Move under the `spring-boot` folder and run these commands

```bash
oc delete --force --grace-period=0 all --all
oc delete pvc/m2-data

go build -o sb *.go
export PATH=$PATH:$(pwd)
export CURRENT=$(pwd)

cd spring-boot

mvn clean package
sb init -n cloud-demo
sb push --mode binary
sb exec start
cd $CURRENT
```

- Execute this command within another terminal

```bash
URL="http://$(oc get routes/spring-boot-http -o jsonpath='{.spec.host}')"
curl $URL/api/greeting
```

## Test 3 : debug

- Log on to an OpenShift cluster with an `admin` role
- Open or create the following project : `cloud-demo`
- Move under the `spring-boot` folder and run these commands

```bash
oc delete --force --grace-period=0 all --all
oc delete pvc/m2-data

go build -o sb *.go
export PATH=$PATH:$(pwd)
export CURRENT=$(pwd)
 
cd spring-boot
sb init -n cloud-demo
sb push --mode binary
sb exec stop
sb debug
cd $CURRENT
```

## Test 4 : source -> compile -> kill pod -> compile again (m2 repo is back again)

- Log on to an OpenShift cluster with an `admin` role
- Open or create the following project : `cloud-demo`
- Move under the `spring-boot` folder and run these commands

```bash
oc delete --force --grace-period=0 all --all
oc delete pvc/m2-data

go build -o sb *.go
export PATH=$PATH:$(pwd)
export CURRENT=$(pwd)

cd spring-boot

sb init -n cloud-demo
sb push --mode source
sb compile
oc delete --grace-period=0 --force=true pod -l app=spring-boot-http 
sb push --mode source
sb compile
sb exec start
cd $CURRENT
```

## Test 5 : build (code not yet finalized as image is build bit no deployment is available)

- Log on to an OpenShift cluster with an `admin` role
- Open or create the following project : `cloud-demo`
- Move under the `spring-boot` folder and run these commands

```bash
oc delete --force --grace-period=0 all --all
oc delete pvc/m2-data

go build -o sb *.go
export PATH=$PATH:$(pwd)
export CURRENT=$(pwd)

cd spring-boot

sb init -n cloud-demo
sb build
cd $CURRENT
```

## Test 6 : Scaffold a project

```bash
go build -o sb *.go
export PATH=$PATH:$(pwd)
export CURRENT=$(pwd)

cd pkg/generator && go run server.go
mkdir -p ~/Temp/springboot-simple && cd  ~/Temp/springboot-simple
sb create -t simple
ls -la
cd $CURRENT
```