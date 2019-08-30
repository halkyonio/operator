# A Real Example

  * [Introduction](#introduction)
  * [Demo's time](#demos-time)
     * [Build the application](#build-the-application)
     * [Install the components on the cluster](#install-the-components-on-the-cluster)
     * [Check if the Component Client is replying](#check-if-the-component-client-is-replying)
     * [Switch from Dev to Build mode](#switch-from-dev-to-build-mode)

## Introduction

This demo implements the example application described in the [Halkyon introduction](../README.md#introduction). We encourage 
you to read it to familiarize yourself with Halkyon and [its concepts](../README.md#key-concepts).

In this demo, we will, moreover, go one step further in the simplification by generating our Halkyon resources based on 
information found in our applications' `application.properties` descriptors. This automatic generation is handled by [`dekorate`](https://dekorate.io).

## Demo time

Build the full `demo` project using `maven`:
```bash
mvn clean package -f demo
``` 

As our Spring Boot maven projects use the `dekorate` halkyon starter

```xml
<dependency>
    <groupId>io.dekorate</groupId>
    <artifactId>halkyon-spring-starter</artifactId>
    <scope>compile</scope>
</dependency>
```

dekorate will generate a `halkyon.yml|json` file under the `target/classes/META-INF/dekorate` during the `package` phase. 
They will be next used to deploy the different `components` on the cluster with the help of the `kubectl` client tool.

```bash
<maven_project>/target/classes/META-INF/dekorate/
```

A few explanation is needed here in order to understand what we did within the different projects to configure them
and to simplify the process to deploy such a complex application on a cluster :-).

As [Dekorate](http://dekorate.io) project supports 2 configuration's mode: `Java Annotation` or `Annotationless using property file`.
When you prefer to use the `Java annotations`, then follow the instructions defined here after.

The `@HalkyonComponent` annotation has been added within the Spring Boot `Application` java class part of each Apache maven project. 

**client**
```java
@HalkyonComponent(
    name = "fruit-client-sb",
    exposeService = true
)
```

When `dekorate` dependency will be called, it will scan the java classes, search about such annotations and if they exist, it will generate from the information provided a `halkyon.yml` file.

Additional information could also be calculated automatically as `Dekorate` supports different frameworks.
When we use Spring Boot Dekorate, then the following parameters `runtime` and `version` will be set respectively to `spring-boot` and `2.1.6.RELEASE`.

The `@HalkyonLink` annotation express,  using an `@Env`, the name of the variable to be injected within the pod in order to let the Spring Boot Application
to configure its `HTTP Client` to access the `HTTP endpoint` exposed by the backend service.
 
**client**
```java
@HalkyonLink(
    name = "link-to-fruit-backend",
    componentName = "fruit-client-sb",
    type = Type.Env,
    envs = @Env(
        name = "ENDPOINT_BACKEND",
        value = "http://fruit-backend-sb:8080/api/fruits"
    )
)
```

Like for the Client's component, we will define a `@HalkyonComponent` and `@HalkyonLink` annotations for the `Backend` component. The link will inject from the secret referenced, the parameters that the application
will use to create a `DataSource`'s java bean able to call the `PostgreSQL` instance.

**Backend**
```java
@HalkyonComponent(
    name = "fruit-backend-sb",
    exposeService = true
)
```

```java
@HalkyonLink(
    name = "link-to-database",
    componentName = "fruit-backend-sb",
    type = Type.Secret,
    ref = "postgresql-db")
```             
                
To configure the postgresql database to be used by the backend, we will use the `@HalkyonCapability` annotation to 
configure the database, the user, password and database name.
                
```java
@HalkyonCapability(
    name = "postgres-db",
    category = "database",
    kind = "postgres",
    version = "10",
    parameters = {
       @Parameter(name = "DB_USER", value = "admin"),
       @Parameter(name = "DB_PASSWORD", value = "admin"),
       @Parameter(name = "DB_NAME", value = "sample-db"),
    }
)
```

If now, as defined within this demo project, you prefer to use the `Annotationless` mode supported by [dekorate](https://github.com/dekorateio/dekorate#annotation-less-configuration), then simply enrich the `application.yml` file of Spring Boot
with the corresponding `halkyon` parameters.

By example the `Backend` component will be defined as such
```yaml
dekorate:
  component:
    name: fruit-backend-sb
    deploymentMode: dev
    exposeService: true
  link:
    name: link-to-database
    componentName: fruit-backend-sb
    type: Secret
    ref: postgres-db-config
  capability:
    name: postgres-db
    category: database
    type: postgres
    version: "10"
    parameters:
    - name: DB_USER
      value: admin
    - name: DB_PASSWORD
      value: admin
    - name: DB_NAME
      value: sample-db
```

To deploy the generated resource files on an existing cluster, execute the following commands:
```bash
kubectl create ns demo
kubectl apply -f fruit-client-sb/target/classes/META-INF/dekorate/halkyon.yml
kubectl apply -f fruit-backend-sb/target/classes/META-INF/dekorate/halkyon.yml
``` 

Wait a few moment and verify if the status of the `components` deployed are ready
```bash
oc get cp -n test
NAME               RUNTIME       VERSION         AGE       MODE      STATUS    MESSAGE   REVISION
fruit-backend-sb   spring-boot   2.1.6.RELEASE   6m        dev       Ready     Ready     
fruit-client-sb    spring-boot   2.1.6.RELEASE   5m        dev       Ready     Ready     

oc get links -n test
NAME                    AGE       STATUS    MESSAGE
link-to-database        6m        Ready     Ready
link-to-fruit-backend   5m        Ready     Ready

oc get capabilities -n test
NAME          CATEGORY   KIND      AGE       STATUS    MESSAGE   REVISION
postgres-db   database             6m        Ready     Ready     
```

Then push the uber jar file within the pod using the following bash script 
```bash
./demo/scripts/k8s_push_start.sh fruit-backend sb demo
./demo/scripts/k8s_push_start.sh fruit-client sb demo
```

Check now if the `Component` Client is replying and calls its `HTTP Endpoint` exposed in order to fetch the `fruits` data from the database consumed by the 
other microservice. The syntax to be used is not the same if the project is running on kubernetes vs openshift as k8s is creating an ingress resource while openshift a route

```bash
# Openshift Route
# export FRONTEND_ROUTE_URL=<service_name>.<hostname_or_ip>.<domain_name>
# oc get route/fruit-client-sb -n demo
export FRONTEND_ROUTE_URL=fruit-client-sb-test.195.201.87.126.nip.io 
curl http://${FRONTEND_ROUTE_URL}/api/client
[{"id":1,"name":"Cherry"},{"id":2,"name":"Apple"},{"id":3,"name":"Banana"}]%  

# Kubernetes Ingress
# export FRONTEND_ROUTE_UR=<CLUSTER_IP>
export FRONTEND_ROUTE_URL=195.201.87.126
curl -H "Host: fruit-client-sb" http://${FRONTEND_ROUTE_URL}/api/client
[{"id":1,"name":"Cherry"},{"id":2,"name":"Apple"},{"id":3,"name":"Banana"}]%  
```

### Switch from Dev to Build mode

Patch the Component which has been deployed and change the deploymentMode from `dev` to `build`. Next, wait till the Halkyon operator will create a
Tekton Build's pod responsible to git clone the project and to perform a s2i build. When the image is pushed to the internal docker registry, then the 
pod of the runtime is created and the service's label will change to use the new service
  
```bash
kubectl patch cp fruit-backend-sb -p '{"spec":{"deploymentMode":"build"}}' --type=merge
``` 

